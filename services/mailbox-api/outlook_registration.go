package main

import (
	"net/http"
	"time"

	"github.com/byte-v-forge/common-lib/envx"
	browserautomationv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/browserautomation/v1"
)

const (
	defaultOutlookResultsDir          = "/app/Results"
	defaultOutlookCommandTimeout      = 180 * time.Second
	defaultOutlookBrowserSessionTTL   = 15 * time.Minute
	defaultOutlookBrowserViewportWide = 1365
	defaultOutlookBrowserViewportHigh = 768
)

type outlookRegistrationConfig struct {
	enabled        bool
	resultsDir     string
	proxyRef       string
	locale         string
	acceptLanguage string
	timezone       string
	userAgent      string
	windowWidth    int
	windowHeight   int
	sessionTTL     time.Duration
	commandTimeout time.Duration
	oauthClientID  string
	oauthRedirect  string
	oauthScopes    []string
	httpTimeout    time.Duration
}

type outlookRegistrationRunner struct {
	cfg           outlookRegistrationConfig
	browserClient browserautomationv1.BrowserAutomationServiceClient
	httpClient    *http.Client
}

type mailboxRecord struct {
	email        string
	password     string
	refreshToken string
	accessToken  string
	source       string
}

type oauthResult struct {
	refreshToken string
	accessToken  string
}

func loadOutlookRegistrationConfig() outlookRegistrationConfig {
	locale := envx.StringDefault("OUTLOOK_REGISTER_AUTOMATION_LOCALE", "en-US")
	return outlookRegistrationConfig{
		enabled:        envx.Bool("OUTLOOK_REGISTER_ENABLED", false),
		resultsDir:     envx.StringDefault("OUTLOOK_REGISTER_RESULTS_DIR", defaultOutlookResultsDir),
		proxyRef:       envx.StringDefault("OUTLOOK_REGISTER_AUTOMATION_PROXY_REF", "outlook"),
		locale:         locale,
		acceptLanguage: envx.StringDefault("OUTLOOK_REGISTER_AUTOMATION_ACCEPT_LANGUAGE", acceptLanguage(locale)),
		timezone:       envx.StringDefault("OUTLOOK_REGISTER_AUTOMATION_TIMEZONE", ""),
		userAgent:      envx.StringDefault("OUTLOOK_REGISTER_AUTOMATION_USER_AGENT", ""),
		windowWidth:    envx.PositiveInt("OUTLOOK_REGISTER_AUTOMATION_WINDOW_WIDTH", defaultOutlookBrowserViewportWide),
		windowHeight:   envx.PositiveInt("OUTLOOK_REGISTER_AUTOMATION_WINDOW_HEIGHT", defaultOutlookBrowserViewportHigh),
		sessionTTL:     envx.PositiveDurationSeconds("OUTLOOK_REGISTER_AUTOMATION_SESSION_TTL_SECONDS", defaultOutlookBrowserSessionTTL),
		commandTimeout: envx.PositiveDurationSeconds("OUTLOOK_REGISTER_AUTOMATION_COMMAND_TIMEOUT_SECONDS", defaultOutlookCommandTimeout),
		oauthClientID:  envx.StringDefault("OUTLOOK_REGISTER_OAUTH_CLIENT_ID", defaultOutlookOAuthClientID),
		oauthRedirect:  envx.StringDefault("OUTLOOK_REGISTER_OAUTH_REDIRECT_URL", defaultOutlookOAuthRedirectURL),
		oauthScopes:    splitScopes(envx.StringDefault("OUTLOOK_REGISTER_OAUTH_SCOPES", defaultOutlookOAuthScopes)),
		httpTimeout:    envx.PositiveDurationSeconds("OUTLOOK_REGISTER_HTTP_TIMEOUT_SECONDS", 30*time.Second),
	}
}

func newOutlookRegistrationRunner(cfg outlookRegistrationConfig, browserClient browserautomationv1.BrowserAutomationServiceClient, httpClient *http.Client) *outlookRegistrationRunner {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.httpTimeout}
	}
	return &outlookRegistrationRunner{
		cfg:           cfg,
		browserClient: browserClient,
		httpClient:    httpClient,
	}
}
