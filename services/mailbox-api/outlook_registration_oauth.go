package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	browserautomationv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/browserautomation/v1"
	"github.com/google/uuid"
)

func (r *outlookRegistrationRunner) runBrowserOAuth(ctx context.Context, email string, password string) (oauthResult, error) {
	session, err := r.startSession(ctx, email)
	if err != nil {
		return oauthResult{}, err
	}
	defer r.stopSession(session)

	state := uuid.NewString()
	authURL := r.oauthAuthorizeURL(state)
	results, err := r.execute(ctx, session, "outlook.oauth", []*browserautomationv1.BrowserCommand{
		navigateCommand("open-oauth", authURL, r.cfg.commandTimeout),
		evaluateCommand("complete-oauth", outlookOAuthScript, map[string]any{
			"email":    email,
			"password": password,
		}, r.cfg.commandTimeout),
	})
	if err != nil {
		return oauthResult{}, err
	}
	resultURL := stringMapValue(commandResultMap(results, "complete-oauth"), "url")
	if resultURL == "" {
		return oauthResult{}, errors.New("OAuth browser flow did not return a redirect URL")
	}
	code, returnedState, oauthErr := oauthCodeFromURL(resultURL)
	if oauthErr != "" {
		return oauthResult{}, fmt.Errorf("OAuth redirect error: %s", safeMailboxText(oauthErr))
	}
	if returnedState != "" && returnedState != state {
		return oauthResult{}, errors.New("OAuth state mismatch")
	}
	if code == "" {
		return oauthResult{}, fmt.Errorf("OAuth code not found in redirect URL: %s", sanitizeURL(resultURL))
	}
	return r.exchangeOAuthCode(ctx, code)
}

func (r *outlookRegistrationRunner) oauthAuthorizeURL(state string) string {
	values := url.Values{}
	values.Set("client_id", r.cfg.oauthClientID)
	values.Set("response_type", "code")
	values.Set("redirect_uri", r.cfg.oauthRedirect)
	values.Set("response_mode", "query")
	values.Set("scope", strings.Join(r.cfg.oauthScopes, " "))
	values.Set("state", state)
	values.Set("prompt", "login")
	return defaultOutlookOAuthAuthorizeURL + "?" + values.Encode()
}

func (r *outlookRegistrationRunner) exchangeOAuthCode(ctx context.Context, code string) (oauthResult, error) {
	values := url.Values{}
	values.Set("client_id", r.cfg.oauthClientID)
	values.Set("scope", strings.Join(r.cfg.oauthScopes, " "))
	values.Set("code", code)
	values.Set("redirect_uri", r.cfg.oauthRedirect)
	values.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, defaultOutlookOAuthTokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return oauthResult{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return oauthResult{}, err
	}
	defer resp.Body.Close()
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return oauthResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return oauthResult{}, fmt.Errorf("OAuth token exchange failed: %s", safeMailboxText(stringMapValue(payload, "error_description")))
	}
	refresh := stringMapValue(payload, "refresh_token")
	access := stringMapValue(payload, "access_token")
	if refresh == "" {
		return oauthResult{}, errors.New("OAuth token exchange returned empty refresh_token")
	}
	return oauthResult{refreshToken: refresh, accessToken: access}, nil
}

func stringMapValue(data map[string]any, key string) string {
	if data == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(data[key]))
}

func oauthCodeFromURL(value string) (code string, state string, oauthErr string) {
	parsed, err := url.Parse(value)
	if err != nil {
		return "", "", safeMailboxError(err)
	}
	query := parsed.Query()
	return strings.TrimSpace(query.Get("code")), strings.TrimSpace(query.Get("state")), strings.TrimSpace(query.Get("error_description"))
}

func splitScopes(value string) []string {
	parts := strings.Fields(strings.ReplaceAll(value, ",", " "))
	if len(parts) == 0 {
		return strings.Fields(defaultOutlookOAuthScopes)
	}
	return parts
}

func sanitizeURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil {
		return "<invalid-url>"
	}
	query := parsed.Query()
	for _, key := range []string{"code", "state", "session_state"} {
		if query.Has(key) {
			query.Set(key, "***")
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
