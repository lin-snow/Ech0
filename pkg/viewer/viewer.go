// Package viewer provides a unified viewer context abstraction.
package viewer

// Context defines the current viewer identity.
type Context interface {
	UserID() string
	TokenScope() string
}
