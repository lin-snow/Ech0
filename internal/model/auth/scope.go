package model

// Token 类型常量，对应 JWT claims 中的 "typ" 字段。
//
//   - session：浏览器用户会话的 access token（短期，前端存 JS 内存）
//   - access：管理面板签发的 API token（长期，面向 CLI/集成/MCP）
//   - refresh：静默刷新专用（长期，存 HttpOnly Cookie）
//
// ParseToken() 仅接受 session/access，ParseRefreshToken() 仅接受 refresh，两者互不混用。
const (
	TokenTypeSession = "session"
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

const (
	ScopeEchoRead       = "echo:read"
	ScopeEchoWrite      = "echo:write"
	ScopeCommentRead    = "comment:read"
	ScopeCommentWrite   = "comment:write"
	ScopeCommentMod     = "comment:moderate"
	ScopeFileRead       = "file:read"
	ScopeFileWrite      = "file:write"
	ScopeConnectRead    = "connect:read"
	ScopeConnectWrite   = "connect:write"
	ScopeProfileRead    = "profile:read"
	ScopeProfileWrite   = "profile:write"
	ScopeAdminSettings  = "admin:settings"
	ScopeAdminUser      = "admin:user"
	ScopeAdminToken     = "admin:token"
	AudiencePublic      = "public-client"
	AudienceCLI         = "cli"
	AudienceIntegration = "integration"
	AudienceMCPRemote   = "mcp-remote"
)

var validScopes = map[string]struct{}{
	ScopeEchoRead:      {},
	ScopeEchoWrite:     {},
	ScopeCommentRead:   {},
	ScopeCommentWrite:  {},
	ScopeCommentMod:    {},
	ScopeFileRead:      {},
	ScopeFileWrite:     {},
	ScopeConnectRead:   {},
	ScopeConnectWrite:  {},
	ScopeProfileRead:   {},
	ScopeProfileWrite:  {},
	ScopeAdminSettings: {},
	ScopeAdminUser:     {},
	ScopeAdminToken:    {},
}

var validAudiences = map[string]struct{}{
	AudiencePublic:      {},
	AudienceCLI:         {},
	AudienceIntegration: {},
	AudienceMCPRemote:   {},
}

func IsValidScope(scope string) bool {
	_, ok := validScopes[scope]
	return ok
}

func IsValidAudience(audience string) bool {
	_, ok := validAudiences[audience]
	return ok
}

func HasAdminScope(scopes []string) bool {
	for _, scope := range scopes {
		switch scope {
		case ScopeAdminSettings, ScopeAdminUser, ScopeAdminToken:
			return true
		}
	}
	return false
}
