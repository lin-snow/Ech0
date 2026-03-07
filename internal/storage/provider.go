package storage

import (
	"context"
	"log"
	"strings"

	"github.com/google/wire"
	virefs "github.com/lin-snow/VireFS"
	"github.com/lin-snow/ech0/internal/config"
)

func ProvideFS() virefs.FS {
	return NewFS(config.Config().Storage)
}

func ProvideURLResolver() URLResolver {
	return NewURLResolver(config.Config().Storage)
}

var (
	FSSet          = wire.NewSet(ProvideFS)
	URLResolverSet = wire.NewSet(ProvideURLResolver)
	ProviderSet    = wire.NewSet(FSSet, URLResolverSet)
)

// NewFS builds a virefs.FS based on the given StorageConfig.
// File classification (images/, audios/, etc.) is handled by VireFS Schema.
func NewFS(cfg config.StorageConfig) virefs.FS {
	schema := NewFileSchema()
	switch strings.ToLower(strings.TrimSpace(cfg.Mode)) {
	case "s3":
		return buildS3FS(cfg, schema)
	default:
		return buildLocalFS(cfg, schema)
	}
}

// NewURLResolver builds a URLResolver based on the given StorageConfig.
// It applies schema.Resolve internally so callers just pass flat keys.
func NewURLResolver(cfg config.StorageConfig) URLResolver {
	schema := NewFileSchema()
	switch strings.ToLower(strings.TrimSpace(cfg.Mode)) {
	case "s3":
		return buildS3URLResolver(cfg, schema)
	default:
		return buildLocalURLResolver(schema)
	}
}

func buildLocalFS(cfg config.StorageConfig, schema *virefs.Schema) virefs.FS {
	root := cfg.DataRoot
	if root == "" {
		root = "data/files"
	}
	fs, err := virefs.NewLocalFS(root,
		virefs.WithCreateRoot(),
		virefs.WithAtomicWrite(),
		virefs.WithLocalKeyFunc(schema.Resolve),
	)
	if err != nil {
		log.Printf("[storage] failed to create local FS: %v, falling back to defaults", err)
		fs, _ = virefs.NewLocalFS("data/files",
			virefs.WithCreateRoot(),
			virefs.WithAtomicWrite(),
			virefs.WithLocalKeyFunc(schema.Resolve),
		)
	}
	return fs
}

func buildLocalURLResolver(schema *virefs.Schema) URLResolver {
	return func(key string) string {
		return "/api/files/" + schema.Resolve(key)
	}
}

func buildS3FS(cfg config.StorageConfig, schema *virefs.Schema) virefs.FS {
	provider := mapProvider(cfg.Provider)

	var opts []virefs.ObjectOption
	if cfg.PathPrefix != "" {
		opts = append(opts, virefs.WithPrefix(strings.Trim(cfg.PathPrefix, "/")+"/"))
	}
	opts = append(opts, virefs.WithObjectKeyFunc(schema.Resolve))

	endpoint := normalizeEndpoint(cfg.Endpoint, cfg.UseSSL)

	fs, err := virefs.NewObjectFSFromConfig(context.Background(), &virefs.S3Config{
		Provider:  provider,
		Endpoint:  endpoint,
		Region:    cfg.Region,
		Bucket:    cfg.BucketName,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
	}, opts...)
	if err != nil {
		log.Printf("[storage] failed to create S3 FS: %v, falling back to local", err)
		return buildLocalFS(cfg, schema)
	}
	return fs
}

func buildS3URLResolver(cfg config.StorageConfig, schema *virefs.Schema) URLResolver {
	prefix := ""
	if cfg.PathPrefix != "" {
		prefix = strings.Trim(cfg.PathPrefix, "/") + "/"
	}

	cdnURL := strings.TrimSpace(cfg.CDNURL)
	if cdnURL != "" {
		if !strings.HasPrefix(strings.ToLower(cdnURL), "http://") &&
			!strings.HasPrefix(strings.ToLower(cdnURL), "https://") {
			protocol := "http"
			if cfg.UseSSL {
				protocol = "https"
			}
			cdnURL = protocol + "://" + cdnURL
		}
		cdnURL = strings.TrimRight(cdnURL, "/")
		return func(key string) string {
			return cdnURL + "/" + prefix + schema.Resolve(key)
		}
	}

	endpoint := normalizeEndpoint(cfg.Endpoint, cfg.UseSSL)
	baseURL := strings.TrimRight(endpoint, "/") + "/" + cfg.BucketName
	return func(key string) string {
		return baseURL + "/" + prefix + schema.Resolve(key)
	}
}

func normalizeEndpoint(endpoint string, useSSL bool) string {
	if endpoint == "" {
		return endpoint
	}
	lower := strings.ToLower(endpoint)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return endpoint
	}
	if useSSL {
		return "https://" + endpoint
	}
	return "http://" + endpoint
}

func mapProvider(raw string) virefs.Provider {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "minio":
		return virefs.ProviderMinIO
	case "r2":
		return virefs.ProviderR2
	default:
		return virefs.ProviderAWS
	}
}
