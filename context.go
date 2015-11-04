package hero

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type responseType int

const (
	responseData responseType = iota
	responseRedirect
)

var (

	//errNotRedirectResponse is returned when Response is not of type ResponseRedirect
	errNotRedirectResponse = errors.New("Not redirect response")
)

type context struct {
	Type responseType

	// StatusCode is the http status code eg 200
	StatusCode int

	// StatusText is any text to be assoiated with the response.
	StatusText string

	// URL is the redirection URL
	URL string

	// Data is a key value pairs to be encoded in the url query.
	Data map[string]interface{}

	// Headers are response headers
	Headers http.Header

	// HasError is true when there was an error
	HasError bool

	// ErrID is the id of the error if HasError is true. This will be searched in the
	//	BaseOauthErrs
	ErrID string

	InternalError error

	RedirectInFragment bool

	Response http.ResponseWriter
}

func newContext(w http.ResponseWriter) *context {
	ctx := &context{
		Type:       responseData,
		StatusCode: http.StatusOK,
		Data:       make(map[string]interface{}),
		Headers:    make(http.Header),
		Response:   w,
	}
	ctx.Headers.Add(
		"Cache-Control",
		"no-cache, no-store, max-age=0, must-revalidate",
	)
	ctx.Headers.Add("Pragma", "no-cache")
	ctx.Headers.Add("Expires", "Fri, 01 Jan 1990 00:00:00 GMT")
	return ctx
}

// SetData stores key value in the context
func (ctx *context) SetData(key string, value interface{}) {
	ctx.Data[key] = value
}

// ClearData deletes all stored key values.
func (ctx *context) ClearData() {
	for k := range ctx.Data {
		delete(ctx.Data, k)
	}
}

func (ctx *context) SetError(id, desc string) {
	ctx.SetErrorURI(id, desc, "", "")
}

func (ctx *context) SetErrorState(id, desc, state string) {
	ctx.SetErrorURI(id, desc, "", state)
}

func (ctx *context) SetRedirect(redirURL string) {
	ctx.Type = responseRedirect
	ctx.URL = redirURL
}

func (ctx *context) SetRedirectFragment(hasFragment bool) {
	ctx.RedirectInFragment = hasFragment
}

func (ctx *context) SetErrorURI(id, desc, uri, state string) {
	if desc == "" {
		desc = baseOauthErrs.Get(id)
	}
	ctx.HasError = true
	ctx.ErrID = id
	if ctx.StatusCode != http.StatusOK {
		ctx.StatusText = desc
	}
	ctx.ClearData()
	ctx.SetData(params.error, id)
	ctx.SetData(params.errDesc, desc)
	ctx.SetData(params.errURI, uri)
	if state != "" {
		ctx.SetData(params.state, state)
	}
}

// GetRedirectURL returns redirect url.
func (ctx *context) GetRedirectURL() (string, error) {
	if ctx.Type != responseRedirect {
		return "", errNotRedirectResponse
	}
	link, err := url.Parse(ctx.URL)
	if err != nil {
		return "", err
	}

	q := link.Query()

	for k, v := range ctx.Data {
		q.Set(k, fmt.Sprint(v))
	}

	link.RawQuery = q.Encode()

	if ctx.RedirectInFragment {
		link.RawQuery = ""
		link.Fragment, err = url.QueryUnescape(q.Encode())
		if err != nil {
			return "", err
		}
	}
	return link.String(), nil
}

func (ctx *context) CommitJSON() error {

	if ctx.InternalError != nil {
		// TODO log this?
	}
	for k, h := range ctx.Headers {
		for _, v := range h {
			ctx.Response.Header().Add(k, v)
		}
	}

	switch ctx.Type {
	case responseRedirect:
		link, err := ctx.GetRedirectURL()
		if err != nil {
			return err
		}
		ctx.Response.Header().Add(params.location, link)
		ctx.Response.WriteHeader(http.StatusFound)
	default:
		ctx.Response.Header().Set("Content-Type", "application/json")
		ctx.Response.WriteHeader(ctx.StatusCode)
		err := json.NewEncoder(ctx.Response).Encode(ctx.Data)
		if err != nil {
			return err
		}
	}
	return nil
}
