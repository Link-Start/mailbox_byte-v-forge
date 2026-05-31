package main

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	absauth "github.com/microsoft/kiota-abstractions-go/authentication"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	msgraphgocore "github.com/microsoftgraph/msgraph-sdk-go-core"
)

type staticAccessTokenProvider struct {
	token     string
	validator *absauth.AllowedHostsValidator
}

func newStaticAccessTokenProvider(token string) (*staticAccessTokenProvider, error) {
	validator, err := absauth.NewAllowedHostsValidatorErrorCheck([]string{"graph.microsoft.com"})
	if err != nil {
		return nil, err
	}
	return &staticAccessTokenProvider{token: token, validator: validator}, nil
}

func (p *staticAccessTokenProvider) GetAuthorizationToken(_ context.Context, _ *url.URL, _ map[string]interface{}) (string, error) {
	return strings.TrimSpace(p.token), nil
}

func (p *staticAccessTokenProvider) GetAllowedHostsValidator() *absauth.AllowedHostsValidator {
	return p.validator
}

func newGraphClient(accessToken string, httpClient *http.Client) (*msgraphsdk.GraphServiceClient, error) {
	tokenProvider, err := newStaticAccessTokenProvider(accessToken)
	if err != nil {
		return nil, err
	}
	authProvider := absauth.NewBaseBearerTokenAuthenticationProvider(tokenProvider)
	options := msgraphsdk.GetDefaultClientOptions()
	graphHTTPClient := msgraphgocore.GetDefaultClient(&options)
	if httpClient != nil && httpClient.Timeout > 0 {
		graphHTTPClient.Timeout = httpClient.Timeout
	}
	adapter, err := msgraphsdk.NewGraphRequestAdapterWithParseNodeFactoryAndSerializationWriterFactoryAndHttpClient(authProvider, nil, nil, graphHTTPClient)
	if err != nil {
		return nil, err
	}
	return msgraphsdk.NewGraphServiceClient(adapter), nil
}
