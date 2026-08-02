// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/job"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	"github.com/lin-snow/ech0/pkg/busen"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const adminID = "user-test-0001"

// TestMain pins a file-free, error-level logger so the job.Manager's goroutine
// logging never lazily initializes the default logger (which would create a
// data/app.log under the package dir). It also keeps test output quiet.
func TestMain(m *testing.M) {
	logUtil.InitLoggerWithConfig(logUtil.LogConfig{Level: "error", File: logUtil.FileConfig{Enable: false}})
	os.Exit(m.Run())
}

// fakeJobRepo is a tiny in-memory, concurrency-safe job.JobRepository. The real
// job.Manager spins a goroutine on Submit success that calls Upsert, so the map
// is guarded by a mutex to stay race-clean.
type fakeJobRepo struct {
	mu     sync.Mutex
	jobs   map[string]jobModel.Job
	getErr error // forced (non-NotFound) error for GetByType, when set
}

func newFakeJobRepo() *fakeJobRepo {
	return &fakeJobRepo{jobs: make(map[string]jobModel.Job)}
}

func (r *fakeJobRepo) Upsert(_ context.Context, j *jobModel.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[j.Type] = *j
	return nil
}

func (r *fakeJobRepo) GetByType(_ context.Context, jobType string) (jobModel.Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.getErr != nil {
		return jobModel.Job{}, r.getErr
	}
	j, ok := r.jobs[jobType]
	if !ok {
		return jobModel.Job{}, job.ErrNotFound
	}
	return j, nil
}

func (r *fakeJobRepo) SweepRunning(_ context.Context, _ string) error { return nil }

func (r *fakeJobRepo) Delete(_ context.Context, jobType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.jobs, jobType)
	return nil
}

func (r *fakeJobRepo) seed(j jobModel.Job) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[j.Type] = j
}

// noopRunner immediately succeeds; used only so Submit's happy path can return a
// pending row. The async transition to success is irrelevant to the assertions.
type noopRunner struct{}

func (noopRunner) Run(context.Context, []byte, job.ReportFunc) (any, error) { return nil, nil }

// newService wires a MigratorService over a mocked CommonService and a real
// job.Manager backed by an in-memory repo. busProvider may be nil-returning for
// methods that don't touch the bus.
func newService(common CommonService, repo *fakeJobRepo, bus *busen.Bus) *MigratorService {
	return NewMigratorService(common, job.NewManager(repo), func() *busen.Bus { return bus })
}

func expectUser(t *testing.T, common *commonmock.MockService, u userModel.User, err error) {
	t.Helper()
	common.EXPECT().CommonGetUserByUserId(mock.Anything, adminID).Return(u, err)
}

func adminUser() userModel.User  { return helpers.NewUser(helpers.AsAdmin) }
func normalUser() userModel.User { return helpers.NewUser() }

// ---------------------------------------------------------------------------
// StartGlobalMigration
// ---------------------------------------------------------------------------

func TestStartGlobalMigration(t *testing.T) {
	validReq := func() migratorModel.StartGlobalMigrationRequest {
		return migratorModel.StartGlobalMigrationRequest{
			SourceType:    migratorModel.MigrationSourceEch0,
			SourcePayload: map[string]any{"tmp_dir": "files/tmp/ech0_x"},
		}
	}

	t.Run("invalid request rejected before auth", func(t *testing.T) {
		common := commonmock.NewMockService(t) // no auth call expected
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.StartGlobalMigration(helpers.CtxAsUser(adminID), migratorModel.StartGlobalMigrationRequest{
			SourceType:    "bogus",
			SourcePayload: map[string]any{"tmp_dir": "files/tmp/x"},
		})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_REQUEST_BODY, err.Error())
	})

	t.Run("non-admin denied", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, normalUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.StartGlobalMigration(helpers.CtxAsUser(adminID), validReq())
		require.Error(t, err)
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})

	t.Run("user lookup error propagates", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		boom := errors.New("db down")
		expectUser(t, common, userModel.User{}, boom)
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.StartGlobalMigration(helpers.CtxAsUser(adminID), validReq())
		assert.ErrorIs(t, err, boom)
	})

	t.Run("no runner registered surfaces submit error and cleans tmp", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		s := newService(common, newFakeJobRepo(), nil) // no Register => ErrNoRunner
		_, err := s.StartGlobalMigration(helpers.CtxAsUser(adminID), validReq())
		require.Error(t, err)
		// Not the already-running message; the raw ErrNoRunner is returned.
		assert.NotEqual(t, "请先结束/清理当前迁移", err.Error())
		assert.ErrorIs(t, err, job.ErrNoRunner)
	})

	t.Run("already running mapped to friendly message", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		repo.seed(jobModel.Job{Type: jobModel.TypeMigration, Status: jobModel.StatusRunning})
		s := newService(common, repo, nil)
		s.jobManager.Register(jobModel.TypeMigration, noopRunner{})
		_, err := s.StartGlobalMigration(helpers.CtxAsUser(adminID), validReq())
		require.Error(t, err)
		assert.Equal(t, "请先结束/清理当前迁移", err.Error())
	})

	t.Run("success returns pending DTO with injected created_by", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		s := newService(common, repo, nil)
		s.jobManager.Register(jobModel.TypeMigration, noopRunner{})

		dto, err := s.StartGlobalMigration(helpers.CtxAsUser(adminID), validReq())
		require.NoError(t, err)
		assert.Equal(t, 1, dto.Version)
		assert.Equal(t, migratorModel.MigrationSourceEch0, dto.SourceType)
		assert.Equal(t, string(jobModel.StatusPending), dto.Status)
		require.NotNil(t, dto.SourcePayload)
		assert.Equal(t, "files/tmp/ech0_x", dto.SourcePayload["tmp_dir"])
		assert.Equal(t, adminID, dto.SourcePayload["created_by"])
	})
}

// ---------------------------------------------------------------------------
// GetGlobalMigrationStatus
// ---------------------------------------------------------------------------

func TestGetGlobalMigrationStatus(t *testing.T) {
	t.Run("non-admin denied", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, normalUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.GetGlobalMigrationStatus(helpers.CtxAsUser(adminID))
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})

	t.Run("no job row synthesizes idle sentinel", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		dto, err := s.GetGlobalMigrationStatus(helpers.CtxAsUser(adminID))
		require.NoError(t, err)
		assert.Equal(t, 1, dto.Version)
		assert.Equal(t, migratorModel.MigrationStatusIdle, dto.Status)
	})

	t.Run("existing job mapped to DTO", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		repo.seed(jobModel.Job{Type: jobModel.TypeMigration, Status: jobModel.StatusSuccess})
		s := newService(common, repo, nil)
		dto, err := s.GetGlobalMigrationStatus(helpers.CtxAsUser(adminID))
		require.NoError(t, err)
		assert.Equal(t, string(jobModel.StatusSuccess), dto.Status)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		boom := errors.New("read fail")
		repo.getErr = boom
		s := newService(common, repo, nil)
		_, err := s.GetGlobalMigrationStatus(helpers.CtxAsUser(adminID))
		assert.ErrorIs(t, err, boom)
	})
}

// ---------------------------------------------------------------------------
// CancelGlobalMigration
// ---------------------------------------------------------------------------

func TestCancelGlobalMigration(t *testing.T) {
	t.Run("non-admin denied", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, normalUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.CancelGlobalMigration(helpers.CtxAsUser(adminID))
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})

	t.Run("no job row is invalid request", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.CancelGlobalMigration(helpers.CtxAsUser(adminID))
		assert.Equal(t, commonModel.INVALID_REQUEST_BODY, err.Error())
	})

	t.Run("terminal job cannot be cancelled", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		repo.seed(jobModel.Job{Type: jobModel.TypeMigration, Status: jobModel.StatusSuccess})
		s := newService(common, repo, nil)
		_, err := s.CancelGlobalMigration(helpers.CtxAsUser(adminID))
		assert.Equal(t, commonModel.INVALID_REQUEST_BODY, err.Error())
	})

	t.Run("running job cancel returns current DTO", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		repo.seed(jobModel.Job{Type: jobModel.TypeMigration, Status: jobModel.StatusRunning})
		s := newService(common, repo, nil)
		dto, err := s.CancelGlobalMigration(helpers.CtxAsUser(adminID))
		require.NoError(t, err)
		assert.Equal(t, string(jobModel.StatusRunning), dto.Status)
	})
}

// ---------------------------------------------------------------------------
// CleanupGlobalMigration
// ---------------------------------------------------------------------------

func TestCleanupGlobalMigration(t *testing.T) {
	t.Run("non-admin denied", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, normalUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		err := s.CleanupGlobalMigration(helpers.CtxAsUser(adminID))
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})

	t.Run("no job row is idempotent no-op", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		assert.NoError(t, s.CleanupGlobalMigration(helpers.CtxAsUser(adminID)))
	})

	t.Run("running job refuses cleanup", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		repo.seed(jobModel.Job{Type: jobModel.TypeMigration, Status: jobModel.StatusPending})
		s := newService(common, repo, nil)
		err := s.CleanupGlobalMigration(helpers.CtxAsUser(adminID))
		require.Error(t, err)
		assert.Equal(t, "迁移进行中，无法清理", err.Error())
	})

	t.Run("terminal job cleans tmp and deletes row", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		repo.seed(jobModel.Job{
			Type:    jobModel.TypeMigration,
			Status:  jobModel.StatusSuccess,
			Payload: `{"source_type":"ech0","source_payload":{"tmp_dir":"files/tmp/ech0_done"}}`,
		})
		s := newService(common, repo, nil)
		require.NoError(t, s.CleanupGlobalMigration(helpers.CtxAsUser(adminID)))
		// row removed
		_, err := repo.GetByType(context.Background(), jobModel.TypeMigration)
		assert.ErrorIs(t, err, job.ErrNotFound)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		boom := errors.New("read fail")
		repo.getErr = boom
		s := newService(common, repo, nil)
		assert.ErrorIs(t, s.CleanupGlobalMigration(helpers.CtxAsUser(adminID)), boom)
	})
}

// ---------------------------------------------------------------------------
// StartExport
// ---------------------------------------------------------------------------

func TestStartExport(t *testing.T) {
	t.Run("non-admin denied", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, normalUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.StartExport(helpers.CtxAsUser(adminID), migratorModel.StartExportRequest{})
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})

	t.Run("no runner surfaces submit error", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.StartExport(helpers.CtxAsUser(adminID), migratorModel.StartExportRequest{})
		require.Error(t, err)
		assert.NotEqual(t, "导出进行中，请稍候", err.Error())
		assert.ErrorIs(t, err, job.ErrNoRunner)
	})

	t.Run("already running mapped to friendly message", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		repo.seed(jobModel.Job{Type: jobModel.TypeExport, Status: jobModel.StatusRunning})
		s := newService(common, repo, nil)
		s.jobManager.Register(jobModel.TypeExport, noopRunner{})
		_, err := s.StartExport(helpers.CtxAsUser(adminID), migratorModel.StartExportRequest{})
		require.Error(t, err)
		assert.Equal(t, "导出进行中，请稍候", err.Error())
	})

	t.Run("success returns pending export DTO", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		s := newService(common, repo, nil)
		s.jobManager.Register(jobModel.TypeExport, noopRunner{})
		dto, err := s.StartExport(helpers.CtxAsUser(adminID), migratorModel.StartExportRequest{})
		require.NoError(t, err)
		assert.Equal(t, 1, dto.Version)
		assert.Equal(t, string(jobModel.StatusPending), dto.Status)
	})
}

// ---------------------------------------------------------------------------
// GetExportStatus
// ---------------------------------------------------------------------------

func TestGetExportStatus(t *testing.T) {
	t.Run("non-admin denied", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, normalUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.GetExportStatus(helpers.CtxAsUser(adminID))
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})

	t.Run("no job row synthesizes idle sentinel", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		dto, err := s.GetExportStatus(helpers.CtxAsUser(adminID))
		require.NoError(t, err)
		assert.Equal(t, migratorModel.MigrationStatusIdle, dto.Status)
	})

	t.Run("existing job parses outcome payload", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		repo.seed(jobModel.Job{
			Type:    jobModel.TypeExport,
			Status:  jobModel.StatusSuccess,
			Payload: `{"file_name":"snap.zip","size":2048}`,
		})
		s := newService(common, repo, nil)
		dto, err := s.GetExportStatus(helpers.CtxAsUser(adminID))
		require.NoError(t, err)
		assert.Equal(t, "snap.zip", dto.FileName)
		assert.Equal(t, int64(2048), dto.Size)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		boom := errors.New("read fail")
		repo.getErr = boom
		s := newService(common, repo, nil)
		_, err := s.GetExportStatus(helpers.CtxAsUser(adminID))
		assert.ErrorIs(t, err, boom)
	})
}

// ---------------------------------------------------------------------------
// CancelExport
// ---------------------------------------------------------------------------

func TestCancelExport(t *testing.T) {
	t.Run("non-admin denied", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, normalUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.CancelExport(helpers.CtxAsUser(adminID))
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})

	t.Run("no job row is invalid request", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.CancelExport(helpers.CtxAsUser(adminID))
		assert.Equal(t, commonModel.INVALID_REQUEST_BODY, err.Error())
	})

	t.Run("terminal job cannot be cancelled", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		repo.seed(jobModel.Job{Type: jobModel.TypeExport, Status: jobModel.StatusFailed})
		s := newService(common, repo, nil)
		_, err := s.CancelExport(helpers.CtxAsUser(adminID))
		assert.Equal(t, commonModel.INVALID_REQUEST_BODY, err.Error())
	})

	t.Run("running job cancel returns current DTO", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		repo.seed(jobModel.Job{Type: jobModel.TypeExport, Status: jobModel.StatusRunning})
		s := newService(common, repo, nil)
		dto, err := s.CancelExport(helpers.CtxAsUser(adminID))
		require.NoError(t, err)
		assert.Equal(t, string(jobModel.StatusRunning), dto.Status)
	})
}

// ---------------------------------------------------------------------------
// DownloadExport
// ---------------------------------------------------------------------------

func newGinCtx(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/migration/export/download", nil)
	return c, w
}

// chdirTemp moves cwd into a throwaway dir so snapshot.LatestPath()'s relative
// "data/files/snapshots" lookup is deterministic and any writes are auto-cleaned.
func chdirTemp(t *testing.T) string {
	t.Helper()
	orig, err := os.Getwd()
	require.NoError(t, err)
	dir := t.TempDir()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })
	return dir
}

func TestDownloadExport(t *testing.T) {
	t.Run("non-admin denied", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, normalUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		c, _ := newGinCtx(t)
		err := s.DownloadExport(c, helpers.CtxAsUser(adminID), "")
		require.Error(t, err)
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})

	t.Run("no snapshot returns guidance error", func(t *testing.T) {
		chdirTemp(t) // empty cwd => no data/files/snapshots
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		c, _ := newGinCtx(t)
		err := s.DownloadExport(c, helpers.CtxAsUser(adminID), "")
		require.Error(t, err)
		assert.Equal(t, "暂无可下载的产物，请先创建导出", err.Error())
	})

	t.Run("streams latest snapshot and emits event", func(t *testing.T) {
		chdirTemp(t)
		snapDir := filepath.Join("data", "files", "snapshots")
		require.NoError(t, os.MkdirAll(snapDir, 0o755))
		content := []byte("PK-fake-zip-bytes")
		require.NoError(t, os.WriteFile(filepath.Join(snapDir, "ech0-snapshot-2026-01-01.zip"), content, 0o644))

		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		bus := helpers.NewTestBus(t)
		s := newService(common, newFakeJobRepo(), bus)

		c, w := newGinCtx(t)
		require.NoError(t, s.DownloadExport(c, helpers.CtxAsUser(adminID), ""))
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/zip", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment;")
		assert.Equal(t, content, w.Body.Bytes())
	})

	// 快照与胶囊各占一个槽位。取错槽位的后果是用户拿到另一种格式的产物——想备份的拿到不含
	// 账号的胶囊，是最坏的一类静默错误，故对两个方向都断言。
	t.Run("format selects the matching artifact slot", func(t *testing.T) {
		chdirTemp(t)
		snapshotBytes := []byte("PK-snapshot")
		capsuleBytes := []byte("PK-capsule")
		require.NoError(t, os.MkdirAll(filepath.Join("data", "files", "snapshots"), 0o755))
		require.NoError(t, os.MkdirAll(filepath.Join("data", "files", "capsules"), 0o755))
		require.NoError(t, os.WriteFile(
			filepath.Join("data", "files", "snapshots", "ech0_snapshot_2026-01-01_00-00-00.zip"), snapshotBytes, 0o644))
		require.NoError(t, os.WriteFile(
			filepath.Join("data", "files", "capsules", "ech0_capsule_2026-01-01_00-00-00.zip"), capsuleBytes, 0o644))

		for _, tc := range []struct {
			format string
			want   []byte
		}{
			{format: "capsule", want: capsuleBytes},
			{format: "snapshot", want: snapshotBytes},
			// 空 format 是加入格式选择之前的客户端行为，必须仍然拿到快照。
			{format: "", want: snapshotBytes},
		} {
			common := commonmock.NewMockService(t)
			expectUser(t, common, adminUser(), nil)
			s := newService(common, newFakeJobRepo(), helpers.NewTestBus(t))
			c, w := newGinCtx(t)
			require.NoError(t, s.DownloadExport(c, helpers.CtxAsUser(adminID), tc.format))
			assert.Equal(t, tc.want, w.Body.Bytes(), "format=%q", tc.format)
			assert.Contains(t, w.Header().Get("Content-Disposition"), "ech0-")
		}
	})

	// 未知格式必须报错而不是静默回落成快照：悄悄给出另一种产物，用户会拿错东西。
	t.Run("unknown format rejected", func(t *testing.T) {
		chdirTemp(t)
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		c, _ := newGinCtx(t)
		require.Error(t, s.DownloadExport(c, helpers.CtxAsUser(adminID), "tarball"))
	})
}

// ---------------------------------------------------------------------------
// UploadSourceZip
// ---------------------------------------------------------------------------

func zipFileHeader(t *testing.T, filename string, payload []byte) *multipart.FileHeader {
	t.Helper()
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, err := mw.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = fw.Write(payload)
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	form, err := multipart.NewReader(body, mw.Boundary()).ReadForm(int64(len(payload)) + 4096)
	require.NoError(t, err)
	files := form.File["file"]
	require.Len(t, files, 1)
	return files[0]
}

func minimalZip(t *testing.T) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	w, err := zw.Create("hello.txt")
	require.NoError(t, err)
	_, err = w.Write([]byte("hi"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

func TestUploadSourceZip(t *testing.T) {
	t.Run("invalid source type rejected first", func(t *testing.T) {
		common := commonmock.NewMockService(t) // no auth call expected
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.UploadSourceZip(helpers.CtxAsUser(adminID), "bogus", nil)
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_REQUEST_BODY, err.Error())
	})

	t.Run("nil file rejected", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.UploadSourceZip(helpers.CtxAsUser(adminID), migratorModel.MigrationSourceEch0, nil)
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_REQUEST_BODY, err.Error())
	})

	t.Run("non-admin denied", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, normalUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.UploadSourceZip(
			helpers.CtxAsUser(adminID),
			migratorModel.MigrationSourceEch0,
			&multipart.FileHeader{Filename: "src.zip"},
		)
		require.Error(t, err)
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})

	t.Run("existing migration row blocks upload", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		repo := newFakeJobRepo()
		repo.seed(jobModel.Job{Type: jobModel.TypeMigration, Status: jobModel.StatusSuccess})
		s := newService(common, repo, nil)
		_, err := s.UploadSourceZip(
			helpers.CtxAsUser(adminID),
			migratorModel.MigrationSourceEch0,
			&multipart.FileHeader{Filename: "src.zip"},
		)
		require.Error(t, err)
		assert.Equal(t, "请先结束/清理当前迁移", err.Error())
	})

	t.Run("non-zip filename rejected", func(t *testing.T) {
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)
		_, err := s.UploadSourceZip(
			helpers.CtxAsUser(adminID),
			migratorModel.MigrationSourceEch0,
			&multipart.FileHeader{Filename: "src.txt"},
		)
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_REQUEST_BODY, err.Error())
	})

	t.Run("success unpacks and returns tmp dir DTO", func(t *testing.T) {
		chdirTemp(t)
		common := commonmock.NewMockService(t)
		expectUser(t, common, adminUser(), nil)
		s := newService(common, newFakeJobRepo(), nil)

		header := zipFileHeader(t, "ech0-export.zip", minimalZip(t))
		resp, err := s.UploadSourceZip(helpers.CtxAsUser(adminID), migratorModel.MigrationSourceEch0, header)
		require.NoError(t, err)
		assert.Equal(t, migratorModel.MigrationSourceEch0, resp.SourceType)
		assert.Contains(t, resp.TmpDir, "files/tmp/ech0_")
		require.NotNil(t, resp.SourcePayload)
		assert.Equal(t, resp.TmpDir, resp.SourcePayload["tmp_dir"])
		// extract dir was created
		_, statErr := os.Stat(filepath.Join("data", filepath.FromSlash(resp.TmpDir)))
		assert.NoError(t, statErr)
	})
}
