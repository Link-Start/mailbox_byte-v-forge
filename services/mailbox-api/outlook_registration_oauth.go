package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	browserautomationv1 "mailboxapi/internal/contracts/browserautomationv1"
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
	cfg := r.oauthConfig()
	return cfg.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("response_mode", "query"),
		oauth2.SetAuthURLParam("prompt", "login"),
	)
}

func (r *outlookRegistrationRunner) exchangeOAuthCode(ctx context.Context, code string) (oauthResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if r.httpClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, r.httpClient)
	}
	cfg := r.oauthConfig()
	token, err := cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("scope", strings.Join(r.cfg.oauthScopes, " ")))
	if err != nil {
		return oauthResult{}, fmt.Errorf("OAuth token exchange failed: %s", safeMailboxText(err.Error()))
	}
	refresh := strings.TrimSpace(token.RefreshToken)
	if refresh == "" {
		return oauthResult{}, errors.New("OAuth token exchange returned empty refresh_token")
	}
	return oauthResult{refreshToken: refresh, accessToken: strings.TrimSpace(token.AccessToken)}, nil
}

func (r *outlookRegistrationRunner) oauthConfig() oauth2.Config {
	return oauth2.Config{
		ClientID:    r.cfg.oauthClientID,
		RedirectURL: r.cfg.oauthRedirect,
		Scopes:      r.cfg.oauthScopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   defaultOutlookOAuthAuthorizeURL,
			TokenURL:  defaultOutlookOAuthTokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}
