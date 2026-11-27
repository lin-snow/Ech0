// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lin-snow/ech0/pkg/virefs"
)

// Config describes one storage instance as it is persisted in storages.config
// and rebuilt on reload. It deliberately belongs to this package: the byte
// layer must not import the file model, so the service maps rows to Config.
type Config struct {
	ID      string
	Driver  string
	Name    string
	Default bool
	Config  map[string]string
}

// Registry owns the storage_id → Store mapping. It is a process-wide snapshot
// swapped atomically on Reload; readers never take the write lock.
type Registry struct {
	mu     sync.RWMutex
	stores map[string]Store
	// publicBase is the CDN origin, when configured, for a storage id. Public
	// bytes are then served from there instead of through this process.
	publicBase map[string]string
	def        string
}

func NewRegistry() *Registry {
	return &Registry{stores: map[string]Store{}, publicBase: map[string]string{}}
}

// Reload builds every configured store and swaps them in only when all of them
// were built. A half-built registry would route new files to an instance that
// does not exist; failing the reload keeps the previous map serving.
func (r *Registry) Reload(ctx context.Context, cfgs []Config) error {
	stores := make(map[string]Store, len(cfgs))
	bases := make(map[string]string, len(cfgs))
	var defaultID string
	for _, cfg := range cfgs {
		if cfg.ID == "" {
			return fmt.Errorf("%w: empty storage id", ErrUnknownStorage)
		}
		if _, dup := stores[cfg.ID]; dup {
			return fmt.Errorf("%w: duplicate storage id %q", ErrUnknownStorage, cfg.ID)
		}
		store, err := OpenDriver(ctx, cfg)
		if err != nil {
			return fmt.Errorf("storage: open %q (%s): %w", cfg.ID, cfg.Driver, err)
		}
		stores[cfg.ID] = store
		if base := strings.TrimRight(strings.TrimSpace(cfg.Config["cdn_url"]), "/"); base != "" {
			bases[cfg.ID] = base
		}
		if cfg.Default {
			defaultID = cfg.ID
		}
	}
	if len(stores) == 0 {
		return fmt.Errorf("%w: no storage configured", ErrUnknownStorage)
	}
	if defaultID == "" {
		ids := make([]string, 0, len(stores))
		for id := range stores {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		defaultID = ids[0]
	}

	r.mu.Lock()
	r.stores = stores
	r.publicBase = bases
	r.def = defaultID
	r.mu.Unlock()
	return nil
}

// For resolves a storage id. An unknown id is a caller bug, not a 404.
func (r *Registry) For(storageID string) (Store, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	store, ok := r.stores[storageID]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownStorage, storageID)
	}
	return store, nil
}

// Default returns the id and store new files are written to.
func (r *Registry) Default() (string, Store, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	store, ok := r.stores[r.def]
	if !ok {
		return "", nil, fmt.Errorf("%w: default %q", ErrUnknownStorage, r.def)
	}
	return r.def, store, nil
}

// PublicURL is where a public object of this storage is served from when a CDN
// origin is configured for it. It is empty when the object must be served
// through the gateway.
func (r *Registry) PublicURL(storageID, objectKey string) string {
	r.mu.RLock()
	base := r.publicBase[storageID]
	r.mu.RUnlock()
	if base == "" || objectKey == "" {
		return ""
	}
	return base + "/" + objectKey
}

// OpenDriver builds one Store from its persisted configuration.
func OpenDriver(ctx context.Context, cfg Config) (Store, error) {
	switch cfg.Driver {
	case DriverLocal:
		root := cfg.Config["root"]
		if root == "" {
			return nil, fmt.Errorf("%w: local driver needs config.root", ErrUnsupported)
		}
		fs, err := virefs.NewLocalFS(root, virefs.WithCreateRoot(), virefs.WithAtomicWrite())
		if err != nil {
			return nil, err
		}
		return newFSStore(fs), nil
	case DriverS3:
		s3cfg, err := s3ConfigFrom(cfg.Config)
		if err != nil {
			return nil, err
		}
		var opts []virefs.ObjectOption
		if prefix := cfg.Config["path_prefix"]; prefix != "" {
			opts = append(opts, virefs.WithPrefix(prefix))
		}
		fs, err := virefs.NewObjectFSFromConfig(ctx, s3cfg, opts...)
		if err != nil {
			return nil, err
		}
		return newFSStore(fs), nil
	case DriverMem:
		return NewMemStore(), nil
	default:
		return nil, fmt.Errorf("%w: driver %q", ErrUnsupported, cfg.Driver)
	}
}

func s3ConfigFrom(config map[string]string) (*virefs.S3Config, error) {
	usePathStyle := false
	if raw := config["use_path_style"]; raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("%w: use_path_style=%q", ErrUnsupported, raw)
		}
		usePathStyle = parsed
	}
	provider := virefs.ProviderAWS
	switch config["provider"] {
	case "minio":
		provider = virefs.ProviderMinIO
	case "r2":
		provider = virefs.ProviderR2
	}
	endpoint := config["endpoint"]
	if endpoint != "" && !hasScheme(endpoint) {
		if config["use_ssl"] == "true" {
			endpoint = "https://" + endpoint
		} else {
			endpoint = "http://" + endpoint
		}
	}
	cfg := &virefs.S3Config{
		Provider:  provider,
		Endpoint:  endpoint,
		Region:    config["region"],
		Bucket:    config["bucket"],
		AccessKey: config["access_key"],
		SecretKey: config["secret_key"],
	}
	if usePathStyle {
		cfg.UsePathStyle = new(true)
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("%w: s3 driver needs config.bucket", ErrUnsupported)
	}
	return cfg, nil
}

func hasScheme(raw string) bool {
	return len(raw) > 7 && (raw[:7] == "http://" || (len(raw) > 8 && raw[:8] == "https://"))
}

// PutTTL bounds a direct upload's signed PUT. It must stay far below the
// pending-row reclaim window so the janitor never deletes an upload that is
// still in flight (§6.3).
const PutTTL = time.Hour
