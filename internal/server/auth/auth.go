// Package auth identifies the caller of the API and the UI and decides what they may do.
//
// A Provider is one authentication strategy. The server asks it who is calling (Identify) and lets it serve its
// own endpoints under /auth/ (Handle); what an identity may do is decided once, in the server's write gate, from
// Identity.Editor, so a new strategy only has to produce identities.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	StrategyAnonymous = "anonymous"
	StrategyOIDC      = "oidc"
)

const (
	RoleEditor = "editor"
	RoleViewer = "viewer"
	// RoleNone is an authenticated caller that may not use resourcerer, such as a user outside allowedGroups.
	RoleNone = "none"
)

type Identity struct {
	User   string
	Groups []string
	// Viewer may read; Editor may also change workloads, settings and recommendations.
	Viewer bool
	Editor bool
	// ExpiresAt is when the session ends; zero when it does not.
	ExpiresAt time.Time
}

func (id Identity) Role() string {
	switch {
	case id.Editor:
		return RoleEditor
	case id.Viewer:
		return RoleViewer
	}
	return RoleNone
}

// Status describes the provider's configuration, for callers to see when it is out of date.
type Status struct {
	// ConfigStale is set when the last read of the configuration failed and the previous one is still in use.
	ConfigStale bool
	ConfigError string
}

type Provider interface {
	Strategy() string
	// Identify returns the caller, or false when the request is not authenticated.
	Identify(r *http.Request) (Identity, bool)
	// Handle serves the strategy's endpoints under /auth/ and returns false for paths it does not own.
	Handle(w http.ResponseWriter, r *http.Request) bool
	Status() Status
	// UsesSessionCookie reports that the strategy recognizes a caller by a cookie the browser attaches on its own.
	// Only those have a cross-site request to defend against: a browser sends a same-site cookie whether or not the
	// page making the request is one we trust, while a header or token a proxy or the caller sets is not something a
	// foreign page can produce. The server requires mutations to prove they came from the UI when this is true.
	UsesSessionCookie() bool
}

// Anonymous trusts every caller with every permission, like Kiali's anonymous strategy.
type Anonymous struct{}

func (Anonymous) Strategy() string { return StrategyAnonymous }

// Anonymous never issues a cookie, so there is no session for another page to borrow.
func (Anonymous) UsesSessionCookie() bool { return false }

func (Anonymous) Identify(*http.Request) (Identity, bool) {
	return Identity{User: "anonymous", Viewer: true, Editor: true}, true
}

func (Anonymous) Handle(http.ResponseWriter, *http.Request) bool { return false }

func (Anonymous) Status() Status { return Status{} }

type contextKey struct{}

func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

func FromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(contextKey{}).(Identity)
	return id, ok
}

// CookieSecure decides the Secure attribute of the cookies auth sets.
type CookieSecure string

const (
	// CookieSecureAuto marks cookies Secure when the request arrived over HTTPS, directly or through a proxy that
	// sets X-Forwarded-Proto. Behind a TLS-terminating proxy, CookieSecureTrue does not depend on that header.
	CookieSecureAuto  CookieSecure = "auto"
	CookieSecureTrue  CookieSecure = "true"
	CookieSecureFalse CookieSecure = "false"
)

func ParseCookieSecure(s string) (CookieSecure, error) {
	switch c := CookieSecure(s); c {
	case CookieSecureAuto, CookieSecureTrue, CookieSecureFalse:
		return c, nil
	}
	return "", fmt.Errorf("got %q", s)
}

func (c CookieSecure) For(r *http.Request) bool {
	switch c {
	case CookieSecureTrue:
		return true
	case CookieSecureFalse:
		return false
	}
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
