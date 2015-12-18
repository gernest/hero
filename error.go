package hero

//
// OAUTH2 error codes
//

//errorsKeys are keys for various ouath 2.0 error messages.
var errorsKeys = struct {
	InvalidRequest          string
	UnauthoredClient        string
	AccessDenied            string
	UnsupportedResponseType string
	InvalidScope            string
	ServerError             string
	TemporalilyUnavailable  string
	UnsupportedGrantType    string
	InvalidGrant            string
	InvalidClient           string
}{
	"invalid_request",
	"unauthorized_client",
	"access_denied",
	"unsupported_response_type",
	"invalid_scope",
	"server_error",
	"temporarily_unavailable",
	"unsupported_grant_type",
	"invalid_grant",
	"invalid_client",
}

//oauthErrors map of oauth2 error codes and descriptions
type oauthErrors map[string]string

// Get returnes description for oauth error code key.
func (o oauthErrors) Get(key string) string {
	if k, ok := o[key]; ok {
		return k
	}
	return key
}

// baseOauthErrs initializes OAuth2 error codes and descriptions.
// http://tools.ietf.org/html/rfc6749#section-4.1.2.1
// http://tools.ietf.org/html/rfc6749#section-4.2.2.1
// http://tools.ietf.org/html/rfc6749#section-5.2
// http://tools.ietf.org/html/rfc6749#section-7.2
var baseOauthErrs = oauthErrors{
	errorsKeys.InvalidRequest:          "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed.",
	errorsKeys.UnauthoredClient:        "The client is not authorized to request a token using this method.",
	errorsKeys.AccessDenied:            "The resource owner or authorization server denied the request.",
	errorsKeys.UnsupportedResponseType: "The authorization server does not support obtaining a token using this method.",
	errorsKeys.InvalidScope:            "The requested scope is invalid, unknown, or malformed.",
	errorsKeys.ServerError:             "The authorization server encountered an unexpected condition that prevented it from fulfilling the request.",
	errorsKeys.TemporalilyUnavailable:  "The authorization server is currently unable to handle the request due to a temporary overloading or maintenance of the server.",
	errorsKeys.UnsupportedGrantType:    "The authorization grant type is not supported by the authorization server.",
	errorsKeys.InvalidGrant:            "The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client.",
	errorsKeys.InvalidClient:           "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method).",
}
