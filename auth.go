package hero

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

type basicAuth struct {
	UserName string
	Password string
}

type bearerAuth struct {
	Code string
}

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
