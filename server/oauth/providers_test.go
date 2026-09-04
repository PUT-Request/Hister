package oauth

import (
	"context"
	"net/url"
	"slices"
	"strings"
	"testing"
)

func TestGoogleUserInfoRequestKeepsTokenOutOfURL(t *testing.T) {
	t.Parallel()

	req, err := newGoogleUserInfoRequest(context.Background(), "access-token")
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.RawQuery != "" {
		t.Errorf("user info URL contains query parameters: %q", req.URL.RawQuery)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer access-token" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer access-token")
	}
}

func TestRedirectURLScopes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		provider   Provider
		configured []string
		want       []string
	}{
		{
			name:     "google defaults",
			provider: GoogleOAuth{AuthURL: "https://auth.example/google"},
			want: []string{
				scopeUserinfoEmail.String(),
				scopeUserinfoProfile.String(),
			},
		},
		{
			name:     "google configured scopes are additional",
			provider: GoogleOAuth{AuthURL: "https://auth.example/google"},
			configured: []string{
				"https://www.googleapis.com/auth/calendar.readonly",
				scopeUserinfoEmail.String(),
			},
			want: []string{
				scopeUserinfoEmail.String(),
				scopeUserinfoProfile.String(),
				"https://www.googleapis.com/auth/calendar.readonly",
			},
		},
		{
			name:       "github configured scopes are additional",
			provider:   GitHubOAuth{AuthURL: "https://auth.example/github"},
			configured: []string{"user:email"},
			want:       []string{scopeReadUser.String(), "user:email"},
		},
		{
			name:       "oidc configured scopes are additional",
			provider:   &OIDCOAuth{AuthURL: "https://auth.example/oidc"},
			configured: []string{"groups", scopeOpenID.String()},
			want: []string{
				scopeOpenID.String(),
				scopeEmail.String(),
				scopeProfile.String(),
				"groups",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			redirectURL := test.provider.GetRedirectURL(NewRedirectURIRequest(
				"client-id",
				"https://hister.example/api/oauth/callback",
				"state-token",
				test.configured,
			))
			u, err := url.Parse(redirectURL)
			if err != nil {
				t.Fatalf("parse redirect URL: %v", err)
			}

			query := u.Query()
			if got := strings.Fields(query.Get(scopeName.String())); !slices.Equal(got, test.want) {
				t.Errorf("scopes = %v, want %v", got, test.want)
			}
			if got := query.Get("client_id"); got != "client-id" {
				t.Errorf("client_id = %q, want %q", got, "client-id")
			}
			if got := query.Get("redirect_uri"); got != "https://hister.example/api/oauth/callback" {
				t.Errorf("redirect_uri = %q", got)
			}
			if got := query.Get("response_type"); got != responseTypeCode.String() {
				t.Errorf("response_type = %q, want %q", got, responseTypeCode)
			}
			if got := query.Get("state"); got != "state-token" {
				t.Errorf("state = %q, want %q", got, "state-token")
			}
		})
	}
}
