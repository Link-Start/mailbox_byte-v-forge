package main

const (
	defaultOutlookOAuthClientID     = "9e5f94bc-e8a4-4e73-b8be-63364c29d753"
	defaultOutlookOAuthRedirectURL  = "https://login.microsoftonline.com/common/oauth2/nativeclient"
	outlookOAuthMailReadScope       = "https://graph.microsoft.com/Mail.Read"
	defaultOutlookOAuthScopes       = "offline_access " + outlookOAuthMailReadScope
	defaultOutlookOAuthAuthorizeURL = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	defaultOutlookOAuthTokenURL     = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
)
