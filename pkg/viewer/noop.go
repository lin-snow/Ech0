package viewer

type NoopViewer struct{}

func NewNoopViewer() *NoopViewer { return &NoopViewer{} }
func (v *NoopViewer) UserID() string {
	return ""
}
func (v *NoopViewer) TokenType() string { return "" }
func (v *NoopViewer) Scopes() []string  { return nil }
func (v *NoopViewer) Audience() []string {
	return nil
}
func (v *NoopViewer) TokenID() string { return "" }

var _ Context = (*NoopViewer)(nil)
