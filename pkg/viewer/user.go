package viewer

type UserViewer struct {
	userID    string
	tokenType string
	scopes    []string
	audience  []string
	tokenID   string
}

func NewUserViewer(userID string) *UserViewer {
	return &UserViewer{userID: userID}
}

func NewUserViewerWithToken(
	userID string,
	tokenType string,
	scopes []string,
	audience []string,
	tokenID string,
) *UserViewer {
	return &UserViewer{
		userID:    userID,
		tokenType: tokenType,
		scopes:    append([]string(nil), scopes...),
		audience:  append([]string(nil), audience...),
		tokenID:   tokenID,
	}
}

func (v *UserViewer) UserID() string    { return v.userID }
func (v *UserViewer) TokenType() string { return v.tokenType }
func (v *UserViewer) Scopes() []string  { return append([]string(nil), v.scopes...) }
func (v *UserViewer) Audience() []string {
	return append([]string(nil), v.audience...)
}
func (v *UserViewer) TokenID() string { return v.tokenID }

var _ Context = (*UserViewer)(nil)
