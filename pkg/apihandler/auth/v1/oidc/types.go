package oidc

// Config ...
type Config struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	ResponseModesSupported            []string `json:"response_modes_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
}

// TokenResponse ...
type TokenResponse struct {
	TokenType        string `json:"token_type"`
	AccessToken      string `json:"access_token"`
	ExpiresIn        uint   `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn uint   `json:"refresh_expires_in"`
	IDToken          string `json:"id_token"`
}

// UserInfo ...
type UserInfo struct {
	Subject  string `json:"sub"`
	UserName string `json:"preferred_username"`
}

// ErrorResponse ...
type ErrorResponse struct {
	ErrorCode   string `json:"error"`
	Description string `json:"error_description"`
	State       string `json:"state"`
}
