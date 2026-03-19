package ech0v3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	virefs "github.com/lin-snow/VireFS"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/migrator/spec"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEch0V3MigrateRegression(t *testing.T) {
	setRepoRootAsWorkingDir(t)

	targetDB := setupMigrationTargetDB(t)
	database.SetDB(targetDB)

	extractor := NewExtractor()
	start := time.Now()
	result, err := extractor.Migrate(context.Background(), spec.MigrateRequest{
		SourcePayload: map[string]any{
			"tmp_dir":           "files/tmp/ech0_v3_019cdc8c-b7ef-7278-9eb9-77185a34184e",
			"created_by":        "migration-test-user",
			"failure_threshold": 0.9,
		},
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	t.Logf("migration elapsed=%s processed=%d success=%d fail=%d", elapsed, result.Processed, result.SuccessCount, result.FailCount)
	if result.Total == 0 {
		t.Fatalf("unexpected total=0")
	}
	if result.Processed != result.SuccessCount+result.FailCount {
		t.Fatalf("invalid processed stats: processed=%d success=%d fail=%d", result.Processed, result.SuccessCount, result.FailCount)
	}
	if result.SuccessCount == 0 {
		t.Fatalf("unexpected success_count=0")
	}
	if result.FailCount == result.Total {
		t.Fatalf("all rows failed unexpectedly")
	}

	var tagCount int64
	if err := targetDB.Model(&echoModel.Tag{}).Count(&tagCount).Error; err != nil {
		t.Fatalf("count tags failed: %v", err)
	}
	if tagCount == 0 {
		t.Fatalf("expected tags migrated, got 0")
	}

	var echoFileCount int64
	if err := targetDB.Model(&fileModel.EchoFile{}).Count(&echoFileCount).Error; err != nil {
		t.Fatalf("count echo_files failed: %v", err)
	}
	if echoFileCount == 0 {
		t.Fatalf("expected echo_files migrated, got 0")
	}
}

type stubFSForPut struct {
	putCalled bool
	putBody   []byte
	putErr    error
}

func (s *stubFSForPut) Get(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (s *stubFSForPut) Put(_ context.Context, _ string, r io.Reader, _ ...virefs.PutOption) error {
	s.putCalled = true
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	s.putBody = b
	return s.putErr
}

func (s *stubFSForPut) Delete(_ context.Context, _ string) error {
	return errors.New("not implemented")
}

func (s *stubFSForPut) List(_ context.Context, _ string) (*virefs.ListResult, error) {
	return nil, errors.New("not implemented")
}

func (s *stubFSForPut) Stat(_ context.Context, _ string) (*virefs.FileInfo, error) {
	return nil, errors.New("not implemented")
}

func (s *stubFSForPut) Access(_ context.Context, _ string) (*virefs.AccessInfo, error) {
	return nil, errors.New("not implemented")
}

func (s *stubFSForPut) Exists(_ context.Context, _ string) (bool, error) {
	return false, errors.New("not implemented")
}

func setRepoRootAsWorkingDir(t *testing.T) {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve test file path failed")
	}
	repoRoot := filepath.Clean(strings.TrimSuffix(filename, "internal/migrator/extractor/ech0v3/extractor_perf_test.go"))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("chdir to repo root failed: %v", err)
	}
}

func setupMigrationTargetDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "migration_target.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open target db failed: %v", err)
	}
	if err := db.AutoMigrate(
		&echoModel.Echo{},
		&echoModel.EchoExtension{},
		&echoModel.Tag{},
		&echoModel.EchoTag{},
		&fileModel.File{},
		&fileModel.EchoFile{},
	); err != nil {
		t.Fatalf("auto migrate target db failed: %v", err)
	}
	return db
}

func TestBuildMigratedFileURL_NormalizeLegacyImagesURL(t *testing.T) {
	url := buildMigratedFileURL("1_1772108750_aa7e4e.jpeg", "/images/1_1772108750_aa7e4e.jpeg", "local")
	if url != "/api/files/images/1_1772108750_aa7e4e.jpeg" {
		t.Fatalf("unexpected migrated url: %s", url)
	}
}

func TestBuildMigratedFileURL_KeepSchemaPathWithoutDoublePrefix(t *testing.T) {
	url := buildMigratedFileURL("images/1_1772108750_aa7e4e.jpeg", "/images/1_1772108750_aa7e4e.jpeg", "local")
	if url != "/api/files/images/1_1772108750_aa7e4e.jpeg" {
		t.Fatalf("unexpected migrated url for schema path key: %s", url)
	}
}

func TestBuildObjectURLFromSetting(t *testing.T) {
	setting := &settingModel.S3Setting{
		Enable:     true,
		Endpoint:   "s3.example.com",
		BucketName: "legacy-bucket",
		PathPrefix: "upload",
		UseSSL:     true,
	}
	url := buildObjectURLFromSetting(*setting, "a.png")
	if url != "https://s3.example.com/legacy-bucket/upload/images/a.png" {
		t.Fatalf("unexpected s3 migrated url: %s", url)
	}
}

func TestBuildObjectURLFromSetting_WithCDN(t *testing.T) {
	// Test CDN with SSL enabled
	setting := &settingModel.S3Setting{
		Enable:     true,
		Endpoint:   "s3.example.com",
		BucketName: "legacy-bucket",
		PathPrefix: "upload",
		UseSSL:     true,
		CDNURL:     "cdn.example.com",
	}
	url := buildObjectURLFromSetting(*setting, "a.png")
	if url != "https://cdn.example.com/upload/images/a.png" {
		t.Fatalf("unexpected s3 migrated url with CDN: %s", url)
	}

	// Test CDN without protocol prefix and SSL disabled
	setting.UseSSL = false
	url = buildObjectURLFromSetting(*setting, "a.png")
	if url != "http://cdn.example.com/upload/images/a.png" {
		t.Fatalf("unexpected s3 migrated url with CDN and no SSL: %s", url)
	}

	// Test CDN that already has https:// prefix
	setting.CDNURL = "https://cdn2.example.com"
	setting.UseSSL = true
	url = buildObjectURLFromSetting(*setting, "a.png")
	if url != "https://cdn2.example.com/upload/images/a.png" {
		t.Fatalf("unexpected s3 migrated url with CDN already having protocol: %s", url)
	}

	// Test CDN that already has http:// prefix (should keep http)
	setting.CDNURL = "http://cdn3.example.com"
	setting.UseSSL = true
	url = buildObjectURLFromSetting(*setting, "a.png")
	if url != "http://cdn3.example.com/upload/images/a.png" {
		t.Fatalf("unexpected s3 migrated url with CDN having http: %s", url)
	}
}

func TestEnsureProtocol(t *testing.T) {
	tests := []struct {
		url    string
		useSSL bool
		want   string
	}{
		{"cdn.example.com", true, "https://cdn.example.com"},
		{"cdn.example.com", false, "http://cdn.example.com"},
		{"https://cdn.example.com", true, "https://cdn.example.com"},
		{"https://cdn.example.com", false, "https://cdn.example.com"},
		{"http://cdn.example.com", true, "http://cdn.example.com"},
		{"http://cdn.example.com", false, "http://cdn.example.com"},
		{"  cdn.example.com  ", true, "https://cdn.example.com"},
	}

	for _, tt := range tests {
		got := ensureProtocol(tt.url, tt.useSSL)
		if got != tt.want {
			t.Errorf("ensureProtocol(%q, %v) = %q, want %q", tt.url, tt.useSSL, got, tt.want)
		}
	}
}

func TestMapSourceS3SettingToV4_Invalid(t *testing.T) {
	_, ok := mapSourceS3SettingToV4(&settingModel.S3Setting{
		Enable:     true,
		Provider:   "r2",
		BucketName: "bucket-only",
	})
	if ok {
		t.Fatalf("expected invalid source s3 setting")
	}
}

func TestNormalizeImageSource(t *testing.T) {
	if got := normalizeImageSource("s3", ""); got != "s3" {
		t.Fatalf("expected s3, got %s", got)
	}
	if got := normalizeImageSource("url", ""); got != "url" {
		t.Fatalf("expected url, got %s", got)
	}
	if got := normalizeImageSource("", "https://cdn.example.com/a.jpg"); got != "url" {
		t.Fatalf("expected url by absolute image_url, got %s", got)
	}
	if got := normalizeImageSource("", "/images/a.jpg"); got != "local" {
		t.Fatalf("expected local fallback, got %s", got)
	}
}

func TestNormalizeExtensionPayload_GithubRawToRepoURL(t *testing.T) {
	payload := normalizeExtensionPayload(
		"GITHUBPROJ",
		map[string]any{"raw": "https://github.com/lin-snow/Ech0"},
		`"https://github.com/lin-snow/Ech0"`,
	)
	if got, _ := payload["repoUrl"].(string); got != "https://github.com/lin-snow/Ech0" {
		t.Fatalf("expected repoUrl mapped, got %#v", payload)
	}
}

func TestNormalizeExtensionPayload_MusicRawToURL(t *testing.T) {
	payload := normalizeExtensionPayload(
		"MUSIC",
		map[string]any{"raw": "https://music.apple.com/cn/song/abc/123"},
		`"https://music.apple.com/cn/song/abc/123"`,
	)
	if got, _ := payload["url"].(string); got == "" {
		t.Fatalf("expected url mapped, got %#v", payload)
	}
}

func TestDeriveS3ObjectKeyMapping(t *testing.T) {
	key, candidates := deriveS3ObjectKeyMapping(
		"1_1773238612_100eca.jpeg",
		"https://minio.vaaat.com/ech0/1_1773238612_100eca.jpeg",
		"",
		"ech0",
	)
	if key != "1_1773238612_100eca.jpeg" {
		t.Fatalf("unexpected mapped key: %s", key)
	}
	foundTargetCandidate := false
	for _, c := range candidates {
		if c == "images/1_1773238612_100eca.jpeg" {
			foundTargetCandidate = true
			break
		}
	}
	if !foundTargetCandidate {
		t.Fatalf("expected schema target candidate in %v", candidates)
	}
}

func TestDeriveS3ObjectKeyMapping_NoSchemaObjectKeyWithBucketURL(t *testing.T) {
	key, candidates := deriveS3ObjectKeyMapping(
		"c0f46d5b6d80a633f7599faa3a421932.png_1759069726",
		"https://minio.vaaat.com/ech0/c0f46d5b6d80a633f7599faa3a421932.png_1759069726",
		"",
		"ech0",
	)
	if key != "c0f46d5b6d80a633f7599faa3a421932.png_1759069726" {
		t.Fatalf("unexpected mapped key: %s", key)
	}
	targetPath := buildObjectStoragePath(settingModel.S3Setting{}, key)
	expected := map[string]bool{
		"c0f46d5b6d80a633f7599faa3a421932.png_1759069726":      false,
		"ech0/c0f46d5b6d80a633f7599faa3a421932.png_1759069726": false,
		targetPath: false,
	}
	for _, c := range candidates {
		if _, ok := expected[c]; ok {
			expected[c] = true
		}
	}
	for k, ok := range expected {
		if !ok {
			t.Fatalf("expected candidate %q in %v", k, candidates)
		}
	}
}

func TestDeriveS3ObjectKeyMapping_RealFailedSamples(t *testing.T) {
	cases := []struct {
		sourceID  string
		objectKey string
		imageURL  string
	}{
		{
			sourceID:  "282",
			objectKey: "c0f46d5b6d80a633f7599faa3a421932.png_1759069726",
			imageURL:  "https://minio.vaaat.com/ech0/c0f46d5b6d80a633f7599faa3a421932.png_1759069726",
		},
		{
			sourceID:  "406",
			objectKey: "1_1773238612_100eca.jpeg",
			imageURL:  "https://minio.vaaat.com/ech0/1_1773238612_100eca.jpeg",
		},
	}
	for _, tc := range cases {
		key, candidates := deriveS3ObjectKeyMapping(tc.objectKey, tc.imageURL, "", "ech0")
		if key == "" {
			t.Fatalf("source_id=%s expected non-empty key", tc.sourceID)
		}
		targetPath := buildObjectStoragePath(settingModel.S3Setting{}, key)
		hasSource := false
		hasSchemaTarget := false
		for _, c := range candidates {
			if c == key || c == "ech0/"+key {
				hasSource = true
			}
			if c == targetPath {
				hasSchemaTarget = true
			}
		}
		if !hasSource || !hasSchemaTarget {
			t.Fatalf("source_id=%s missing required candidates source=%v target=%v all=%v", tc.sourceID, hasSource, hasSchemaTarget, candidates)
		}
	}
}

func TestParseObjectPathFromImageURL(t *testing.T) {
	full, stripped := parseObjectPathFromImageURL("https://minio.vaaat.com/ech0/images/a.jpeg", "ech0")
	if full != "ech0/images/a.jpeg" {
		t.Fatalf("unexpected full path: %s", full)
	}
	if stripped != "images/a.jpeg" {
		t.Fatalf("unexpected stripped path: %s", stripped)
	}
}

func TestBuildObjectStoragePath(t *testing.T) {
	path := buildObjectStoragePath(settingModel.S3Setting{PathPrefix: ""}, "1_1773238612_100eca.jpeg")
	if path != "images/1_1773238612_100eca.jpeg" {
		t.Fatalf("unexpected object storage path: %s", path)
	}
}

func TestNormalizeS3Endpoint(t *testing.T) {
	if got := normalizeS3Endpoint("minio.vaaat.com", true); got != "https://minio.vaaat.com" {
		t.Fatalf("unexpected normalized endpoint: %s", got)
	}
	if got := normalizeS3Endpoint("http://minio.vaaat.com/", true); got != "http://minio.vaaat.com" {
		t.Fatalf("unexpected normalized endpoint with scheme: %s", got)
	}
}

func TestNormalizeS3Region(t *testing.T) {
	if got := normalizeS3Region("minio", "auto"); got != "us-east-1" {
		t.Fatalf("expected minio auto -> us-east-1, got %s", got)
	}
	if got := normalizeS3Region("r2", "auto"); got != "auto" {
		t.Fatalf("expected r2 auto unchanged, got %s", got)
	}
}

func TestPutObjectFromReadCloser(t *testing.T) {
	fs := &stubFSForPut{}
	src := io.NopCloser(bytes.NewBufferString("abc123"))
	if err := putObjectFromReadCloser(fs, "images/a.jpg", src); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fs.putCalled {
		t.Fatalf("expected put called")
	}
	if string(fs.putBody) != "abc123" {
		t.Fatalf("unexpected put body: %q", string(fs.putBody))
	}
}
