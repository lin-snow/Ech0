# 后端测试指南

本文约定 Ech0 后端（Go）的单元测试体系与写法。前端测试见 `web/`（vitest），不在此列。

## 工具与依赖

- **断言**：`github.com/stretchr/testify`（`require` 致命 / `assert` 非致命）。
- **mock**：`github.com/vektra/mockery/v3`（testify 模板）。mockery 只是代码生成器，**不进 `go.mod`**——通过 `go run github.com/vektra/mockery/v3@<pin>` 调用，版本在 `justfile` 的 `MOCKERY_VERSION` 固定，保证任何机器/CI 生成结果一致。
- **不重写既有 stdlib 测试**：历史上用裸 `t.Fatalf` 的优质测试（如 `internal/handler/auth/auth_handler_test.go`）保持原样；新测试统一用 testify。

## 目录结构

```
internal/test/helpers/            # 共享脚手架（手写，package helpers）
  db.go        NewTestDB(t) / NewTestDBWithVec(t)：唯一 DSN 内存 sqlite + SetDB + MigrateDB + t.Cleanup
  viewer.go    CtxAsUser / CtxAsToken / CtxAnonymous：注入身份到 context
  config.go    SetJWTSecret(t, secret)：覆写并自动还原 JWT 密钥
  envelope.go  ParseResult / DecodeData：解析 commonModel.Result 封套，断言 i18n 错误契约
  fixtures.go  NewUser / NewEcho（+ AsAdmin / AsOwner / AsPrivate option）
internal/test/mocks/<domain>mock/ # mockery 生成（mocks.go），如 commentmock.NewMockRepository(t)
.mockery.yaml                     # mockery 配置（集中输出 + SPDX boilerplate）
```

## 包归属规则（关键，避免 import cycle）

- **测试未导出函数**（纯函数，如 `isSafeTagName`、`signFormToken`、`normalizeEchoExtension`）→ 写 **in-package** 测试（`package service`），无需 mock。
- **测试 service 方法 / handler**（需要 mock 协作者）→ 写 **外部测试包**（`package service_test`），`import` 对应的 `internal/test/mocks/<domain>mock`。外部测试包不属于被测包，故 `service_test → commentmock → model/comment` 不成环。

## 写法约定

- **表驱动 + 子测试**：`for _, tc := range cases { t.Run(tc.name, ...) }`；helper 函数首行 `t.Helper()`。
- **错误路径断言 i18n 契约**：用 `helpers.ParseResult` 取 `error_code` / `message_key`，参照 `auth_handler_test.go`。
- **安全/回归测试绑定来源**：注释里写明 `GHSA-xxxx` 或 issue 号（仓库既有惯例）。
- **禁止**：`time.Sleep` 等异步、真实网络、硬编码端口。异步用 channel/同步钩子，HTTP 用 `httptest`。
- **DB 测试**：用 `helpers.NewTestDB(t)`（embedding/vec0 用 `NewTestDBWithVec`）；因依赖全局 `database.SetDB` 单例，**不要 `t.Parallel()`**。
- **SPDX 头**：每个 `.go`（含 `_test.go`）都要有许可证头，`just spdx` 可补齐，CI `just spdx-check` 强制。
- 生成的 mock 带 `// Code generated ... DO NOT EDIT.`，golangci-lint 自动跳过，**勿手改**。

## 示例

### 1) 纯函数（in-package，无 mock）

```go
package service

func TestIsSafeTagName(t *testing.T) {
	cases := []struct{ name, in string; want bool }{
		{"plain", "golang", true},
		{"angle-bracket", "<script>", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isSafeTagName(tc.in))
		})
	}
}
```

### 2) service 方法（外部测试包 + mock）

```go
package service_test

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	commentmock "github.com/lin-snow/ech0/internal/test/mocks/commentmock"
	"github.com/lin-snow/ech0/internal/test/helpers"
)

func TestCreateIntegrationComment_RequiresApproval(t *testing.T) {
	repo := commentmock.NewMockRepository(t) // NewMockXxx(t) 会在 Cleanup 自动校验期望
	repo.EXPECT().
		CreateComment(mock.Anything, mock.Anything).
		Return(nil).Once()

	svc := commentService.NewCommentService(repo /* , 其余 mock 协作者 */)
	err := svc.CreateIntegrationComment(helpers.CtxAsUser("u-1"), /* dto */)
	require.NoError(t, err)
}
```

### 3) DB 集成（仓储层）

```go
func TestEchoRepository_Create(t *testing.T) {
	db := helpers.NewTestDB(t) // 已 SetDB + 建表；t.Cleanup 自动还原
	repo := repository.NewEchoRepository(/* deps */)
	require.NoError(t, repo.CreateEcho(context.Background(), &echo))
	var got echoModel.Echo
	require.NoError(t, db.First(&got, "id = ?", echo.ID).Error)
}
```

## 新增一个 mock

1. 在 `.mockery.yaml` 对应 package 下加接口名（或新增 package 块，指定 `config.dir` / `config.pkgname`）。
2. `just mocks` 重新生成，`just mocks-check` 确认无漂移。
3. 提交生成的 `internal/test/mocks/<domain>mock/mocks.go`。

> 并行写测试时**不要各自改 `.mockery.yaml`**——所需 mock 应一次性在基建阶段生成好，避免冲突。

## 常用命令

```bash
just test         # go test ./...
just test-race    # CGO_ENABLED=1 go test -race ./...
just test-cover   # 覆盖率 + 打印总数
just mocks        # 重新生成 mock
just mocks-check  # mock 漂移即失败
```

## CI

`.github/workflows/test.yml`（PR + push main 触发）：`just mocks-check` → `CGO_ENABLED=1 go test -race -coverprofile`。
覆盖率写入 job summary、上传 artifact，**report-only 不阻断合并**。需要 CGO（sqlite + race），ubuntu-latest 自带 gcc，无需 musl/zig。
