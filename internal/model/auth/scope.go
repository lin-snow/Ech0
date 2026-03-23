package model

const (
	TokenTypeSession = "session"
	TokenTypeAccess  = "access"
)

const (
	ScopeEchoRead       = "echo:read"
	ScopeEchoWrite      = "echo:write"
	ScopeCommentRead    = "comment:read"
	ScopeCommentWrite   = "comment:write"
	ScopeCommentMod     = "comment:moderate"
	ScopeFileRead       = "file:read"
	ScopeFileWrite      = "file:write"
	ScopeProfileRead    = "profile:read"
	ScopeAdminSettings  = "admin:settings"
	ScopeAdminUser      = "admin:user"
	ScopeAdminToken     = "admin:token"
	AudiencePublic      = "public-client"
	AudienceCLI         = "cli"
	AudienceIntegration = "integration"
)

var validScopes = map[string]struct{}{
	ScopeEchoRead:      {},
	ScopeEchoWrite:     {},
	ScopeCommentRead:   {},
	ScopeCommentWrite:  {},
	ScopeCommentMod:    {},
	ScopeFileRead:      {},
	ScopeFileWrite:     {},
	ScopeProfileRead:   {},
	ScopeAdminSettings: {},
	ScopeAdminUser:     {},
	ScopeAdminToken:    {},
}

var validAudiences = map[string]struct{}{
	AudiencePublic:      {},
	AudienceCLI:         {},
	AudienceIntegration: {},
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
