package viewer

type UserViewer struct {
	userID     string
	tokenScope string
}

func NewUserViewer(userID string) *UserViewer { return &UserViewer{userID: userID} }

func NewScopedUserViewer(userID, tokenScope string) *UserViewer {
	return &UserViewer{userID: userID, tokenScope: tokenScope}
}

func (v *UserViewer) UserID() string     { return v.userID }
func (v *UserViewer) TokenScope() string { return v.tokenScope }

var _ Context = (*UserViewer)(nil)
