package hero

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

//basicAuth stores information for basic authentication of oauth clients. Basic authentication
// involves username and password.
type basicAuth struct {
	UserName string
	Password string
}

//bererAuth stores the code for authenciating clients using bearer token.
type bearerAuth struct {
	Code string
}

//getClientAuth returns the basic authentication details from the given request. if the allowParams is
// set to true then the basic auth information will be extracted from the request query parameters.
// Make sure you call r.Parse() before calling this, so as to make the query params available in r.Form
//
// Default the details are extracted from the request header.
func getCLientAuth(r *http.Request, allowQueryParams bool) (*basicAuth, error) {
	auth := &basicAuth{
		UserName: r.Form.Get("client_id"),
		Password: r.Form.Get("client_secret"),
	}
	if allowQueryParams && auth.Password != "" && auth.UserName != "" {
		return auth, nil
	}
	return checkBasicAuth(r)
}

// checkBasiAuth returns basic client athentication  details from the given request. The information is
// extracted from the request header.
func checkBasicAuth(r *http.Request) (*basicAuth, error) {
	var (
		basic                   = "Basic"
		authorize               = "Authorization"
		errInvalidUthorizeHader = errors.New("Invalid authorization header")
	)
	authHeader := r.Header.Get(authorize)
	components := strings.SplitN(authHeader, " ", 2)

	if len(components) != 2 || components[0] != basic {
		return nil, errInvalidUthorizeHader
	}

	base, err := base64.StdEncoding.DecodeString(components[1])
	if err != nil {
		return nil, err
	}

	keyPairs := strings.SplitN(string(base), ":", 2)
	if len(keyPairs) != 2 {
		return nil, errInvalidUthorizeHader
	}
	return &basicAuth{keyPairs[0], keyPairs[1]}, nil

}

//checkBearerAuth checks for bearer token in the request header.
func checkBearerAuth(r *http.Request) *bearerAuth {
	var (
		auth   = "Authorization"
		code   = "code"
		bearer = "Bearer"
	)

	authHeader := r.Header.Get(auth)
	authCode := r.Form.Get(code)
	if authHeader == "" && authCode == "" {
		return nil
	}
	if authHeader != "" {
		components := strings.SplitN(authHeader, " ", 2)
		if (len(components) != 2 || components[0] != bearer) && authCode == "" {
			return nil
		}
		authCode = components[1]
	}
	return &bearerAuth{authCode}
}
