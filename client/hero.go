package client

import (
	"net/http"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/common"
	"github.com/stretchr/gomniauth/oauth2"
	"github.com/stretchr/objx"
)

const (
	heroDefaultScope = "user"
)

var (
	heroProviderName   = "hero"
	HeroProviderDsplay = "Hero"
)

type Config struct {
	ProviderName        string
	ProviderDisplayName string
	AuthURL             string
	TokenURL            string
	ProfileURL          string
	CLientID            string
	CLientSecret        string
	DefaultScope        string
	RedirectURL         string
}

type HeroProvider struct {
	config *common.Config
	cfg    *Config
	trip   common.TripperFactory
}

func New(cfg *Config) *HeroProvider {
	p := &HeroProvider{}
	p.config = &common.Config{Map: objx.MSI(
		oauth2.OAuth2KeyAuthURL, cfg.AuthURL,
		oauth2.OAuth2KeyTokenURL, cfg.TokenURL,
		oauth2.OAuth2KeyClientID, cfg.CLientID,
		oauth2.OAuth2KeySecret, cfg.CLientSecret,
		oauth2.OAuth2KeyRedirectUrl, cfg.RedirectURL,
		oauth2.OAuth2KeyScope, cfg.DefaultScope,
		oauth2.OAuth2KeyAccessType, oauth2.OAuth2AccessTypeOnline,
		oauth2.OAuth2KeyApprovalPrompt, oauth2.OAuth2ApprovalPromptAuto,
		oauth2.OAuth2KeyResponseType, oauth2.OAuth2KeyCode)}
	p.cfg = cfg
	return p
}

// TripperFactory gets an OAuth2TripperFactory
func (h *HeroProvider) TripperFactory() common.TripperFactory {
	if h.trip == nil {
		h.trip = &oauth2.OAuth2TripperFactory{}
	}
	return h.trip
}

// PublicData gets a public readable view of this provider.
func (h *HeroProvider) PublicData(options map[string]interface{}) (interface{}, error) {
	return gomniauth.ProviderPublicData(h, options)
}

// Name is the unique name for this provider.
func (h *HeroProvider) Name() string {
	return h.cfg.ProviderName
}

// DisplayName is the human readable name for this provider.
func (h *HeroProvider) DisplayName() string {
	return h.cfg.ProviderDisplayName
}

// GetBeginAuthURL gets the URL that the client must visit in order
// to begin the authentication process.
//
// The state argument contains anything you wish to have sent back to your
// callback endpoint.
// The options argument takes any options used to configure the auth request
// sent to the provider. In the case of OAuth2, the options map can contain:
//   1. A "scope" key providing the desired scope(s). It will be merged with the default scope.
func (h *HeroProvider) GetBeginAuthURL(state *common.State, options objx.Map) (string, error) {
	if options != nil {
		scope := oauth2.MergeScopes(options.Get(oauth2.OAuth2KeyScope).Str(), h.cfg.DefaultScope)
		h.config.Set(oauth2.OAuth2KeyScope, scope)
	}
	return oauth2.GetBeginAuthURLWithBase(h.config.Get(oauth2.OAuth2KeyAuthURL).Str(), state, h.config)
}

// Get makes an authenticated request and returns the data in the
// response as a data map.
func (h *HeroProvider) Get(creds *common.Credentials, endpoint string) (objx.Map, error) {
	return oauth2.Get(h, creds, endpoint)
}

// GetUser uses the specified common.Credentials to access the users profile
// from the remote provider, and builds the appropriate User object.
func (h *HeroProvider) GetUser(creds *common.Credentials) (common.User, error) {

	profileData, err := h.Get(creds, h.cfg.ProfileURL)

	if err != nil {
		return nil, err
	}

	// build user
	user := NewUser(profileData, creds, h)

	return user, nil
	return nil, nil
}

// CompleteAuth takes a map of arguments that are used to
// complete the authorisation process, completes it, and returns
// the appropriate Credentials.
func (h *HeroProvider) CompleteAuth(data objx.Map) (*common.Credentials, error) {
	return oauth2.CompleteAuth(h.TripperFactory(), data, h.config, h)
}

// GetClient returns an authenticated http.Client that can be used to make requests to
// protected Github resources
func (h *HeroProvider) GetClient(creds *common.Credentials) (*http.Client, error) {
	return oauth2.GetClient(h.TripperFactory(), creds, h)
}
