// Package hero is a heroic oauth2 provider.
package hero

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	// load mysql driver.
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	// loag postgres driver
	_ "github.com/lib/pq"
)

var (

	// requestType contains oauth2 request_type details.
	// that is code, and token.
	requestType = struct {
		Code, Token string
	}{"code", "token"}

	// gratType contains fields that stores oauth2 grant types.
	grantType = struct {
		AuthorizationCode, RefreshToken string
		Password, ClientCredentials     string
		Assertion, Implicit             string
	}{
		"authorization_code", "refresh_token",
		"password", "client_credentials",
		"assertion", "__implicit",
	}

	// params contains varions keys used by hero
	params = struct {
		error         string
		errDesc       string
		errURI        string
		state         string
		grantType     string
		location      string
		clientID      string
		clientSecret  string
		accessToken   string
		tokenType     string
		expiresIn     string
		refreshToken  string
		scope         string
		redirectURL   string
		code          string
		assertion     string
		assertionType string
		responseType  string
	}{
		"error",
		"error_description",
		"error_uri",
		"state",
		"grant_type",
		"Location",
		"client_id",
		"client_secret",
		"access_token",
		"token_type",
		"expires_in",
		"refresh_token",
		"scope",
		"redirect_url",
		"code",
		"assertion",
		"assertion_type",
		"response_type",
	}

	// registerParams contains registration parameters
	registerParams = struct {
		username string
		password string
		confirm  string
		email    string
	}{
		"register_username",
		"register_password",
		"register_confirm",
		"register_email",
	}

	//loginParams conains login parameters
	loginParams = struct {
		username string
		password string
	}{
		"login_username",
		"login_password",
	}

	// contextParams contains keys for context data sent to templates.
	contextParams = struct {
		Config     string
		Message    string
		StatusCode string
	}{
		"Config",
		"Message",
		"StatusCode",
	}
)

const (

	//LoginPath is the route for login.
	LoginPath = "/login"

	// LogoutPath is the route for logout.
	LogoutPath = "/logout"

	//RegisterPath is the route for registration
	RegisterPath = "/register"

	//ProfilePath is the route for user profile
	ProfilePath = "/profile"

	//ClientsPath is the route for user clients
	ClientsPath = "/clients"

	//HomePath is the home page route
	HomePath = "/"

	//StaticPath is the path for static assets.
	StaticPath = "/static/"

	// FlashKey is the key used to store flash messages in seesion
	FlashKey = "_flash"

	csrfTokenLength = 32
)

//Config conains configuration settings for hero.
type Config struct {
	RedirSeparator      string   `json:"redirect_separator"`
	AuthorizationExpire int64    `json:"authorization_expire"`
	AccessExpire        int64    `json:"access_expire"`
	AllowGetAccess      bool     `json:"allow_get_access"`
	AllowedAccessType   []string `json:"allowed_access_type"`
	TokenType           string   `json:"token_type"`
	ProviderName        string   `json:"provider_name"`
	AuthEndpoint        string   `json:"auth_endpoint"`
	TokenEndpoint       string   `json:"token_endpoint"`
	InfoEndpoint        string   `json:"info_endpoint"`
	Port                int      `json:"port"`
	DatabaseDialect     string   `json:"database_dialect"`
	DatabaseConnection  string   `json:"database_connection"`
	TemplatesDir        string   `json:"templates_dir"`
	StaticDir           string   `json:"static_dir"`
	SessionPath         string   `json:"session_path"`
	SessionMaxAge       int      `json:"session_max_age"`
	SessionDomain       string   `json:"session_domain"`
	SessionSecure       bool     `json:"session_secure"`
	SessionHTTPOnly     bool     `json:"session_hhhponly"`
	SessionName         string   `json:"session_name"`
	LoginTemplate       string   `json:"Login_template"`
	ErrorTemplate       string   `json:"error_template"`
	RegisterTemplate    string   `json:"register_template"`
	CLientTemplate      string   `json:"client_template"`
	ProfileTemplate     string   `json:"profile_template"`
	HomeTemplate        string   `json:"home_template"`
	DocsDir             string   `json:"docs_dir"`
	CsrfSecret          string   `json:"csrf_secret"`
}

// AccessAllowed returns true if accesType is allowed.
func (c *Config) AccessAllowed(acessType string) bool {
	var found bool
	for _, k := range c.AllowedAccessType {
		if k == acessType {
			found = true
			break
		}
	}
	return found
}

// GetDoc reads the content of a file named name found inside the *Config.DocsDir
// directory. This is a convenience
// to be used in templates especially inserting contents from markdown files,
// used in combination with other tempalates functions.
func (c *Config) GetDoc(name string) string {
	if c.DocsDir == "" {
		c.DocsDir = "docs"
	}
	fName := filepath.Join(c.DocsDir, name)
	b, err := ioutil.ReadFile(fName)
	if err != nil {
		// Log this?
		return err.Error()
	}
	return string(b)
}

//DefaultConfig returns *Config with default values.
func DefaultConfig() *Config {
	return &Config{
		AllowedAccessType: []string{
			"authorization_code", "refresh_token",
			"password", "client_credentials",
			"assertion",
		},
		TokenType:           "Bearer",
		AuthorizationExpire: 200,
		AccessExpire:        200,
		AuthEndpoint:        "/authorize",
		TokenEndpoint:       "/tokens",
		InfoEndpoint:        "/info",
		DatabaseConnection:  "postgres://postgres:postgres@localhost/hero_test?sslmode=disable",
		DatabaseDialect:     "postgres",
		ErrorTemplate:       "error.html",
		LoginTemplate:       "login.html",
		RegisterTemplate:    "register.html",
		CLientTemplate:      "client.html",
		ProfileTemplate:     "profile.html",
		HomeTemplate:        "home.html",
		TemplatesDir:        "views",
		SessionPath:         "/",
		SessionName:         "_hero",
		Port:                8090,
		CsrfSecret:          "w4PYxQjVP9ZStjWpBt5t28CEBmRs8NPx",
	}
}

// Server is the oauth 2.0  provider.
//
// It implements http.Server interface. This uses *http.ServeMux to register
// and serve its routes.
//
// This provide both resource owner, resource server and authorization server.
type Server struct {
	q     *query
	cfg   *Config
	gen   TokenGenerator
	view  View
	log   Logger
	store *Store
	mux   *mux.Router
}

//NewServer creates a new *Server.
//
//It panics if database connection cannot be established.If view is nil the default
// view is used instead, it panics when the defalut view can not be created.
//
//*Server.Init is called, meaning the returned *Server has all the http endpoints registered,
// the returned instance is ready to be hooked on *http.ListenAndServe.
func NewServer(cfg *Config, gen TokenGenerator, view View) *Server {
	db, err := gorm.Open(cfg.DatabaseDialect, cfg.DatabaseConnection)
	if err != nil {
		panic(err)
	}
	q := &query{}
	q.DB = db
	if view == nil {
		view, err = NewDefaultView(cfg.TemplatesDir, false)
		if err != nil {
			panic(err)
		}
	}
	s := &Server{
		q:     q,
		cfg:   cfg,
		gen:   gen,
		view:  view,
		log:   NewLogger(),
		mux:   mux.NewRouter(),
		store: DefaultStore(q.DB),
	}
	return s.Init()
}

// Init registers the url routes. This uses *http.ServerMux as its router.
func (s *Server) Init() *Server {

	// normal stuffs
	s.mux.HandleFunc(HomePath, s.Home)
	s.mux.HandleFunc(RegisterPath, s.Register)
	s.mux.HandleFunc(LoginPath, s.Login)
	s.mux.HandleFunc(LogoutPath, s.Logout)
	s.mux.HandleFunc(ProfilePath, s.Profile)
	s.mux.HandleFunc(ClientsPath, s.Client)

	// oauth stuffs
	s.mux.HandleFunc(s.cfg.AuthEndpoint, s.Authorize)
	s.mux.HandleFunc(s.cfg.TokenEndpoint, s.Access)
	s.mux.HandleFunc(s.cfg.InfoEndpoint, s.Info)

	// static stuffs
	s.mux.PathPrefix(StaticPath).
		Handler(http.StripPrefix(StaticPath, http.FileServer(http.Dir(s.cfg.StaticDir))))
	return s
}

// Authorize provide oauth2 authorization.
func (s *Server) Authorize(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	ctx := newContext(w)
	redirectURI, err := url.QueryUnescape(r.Form.Get(params.redirectURL))
	if err != nil {
		ctx.SetErrorState(errorsKeys.InvalidRequest, "", "")
		ctx.InternalError = err
		_ = ctx.CommitJSON()
		return
	}

	state := r.Form.Get(params.state)
	scope := r.Form.Get(params.scope)
	clientID := r.Form.Get(params.clientID)

	client, err := s.q.ClientByCode(clientID)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			ctx.SetErrorState(errorsKeys.UnauthorizedClient, "", state)
		} else {
			ctx.SetErrorState(errorsKeys.ServerError, "", state)
		}
		ctx.InternalError = err
		_ = ctx.CommitJSON()
		return
	}

	if client.RedirectURL == "" {
		ctx.SetErrorState(errorsKeys.UnauthorizedClient, "", state)
		_ = ctx.CommitJSON()
		return
	}

	if redirectURI == "" && firstURI(client.RedirectURL, s.cfg.RedirSeparator) == client.RedirectURL {
		redirectURI = firstURI(client.RedirectURL, s.cfg.RedirSeparator)
	}

	if err = validateURIList(client.RedirectURL, redirectURI, s.cfg.RedirSeparator); err != nil {
		ctx.SetErrorState(errorsKeys.InvalidRequest, "", state)
		ctx.InternalError = err
		_ = ctx.CommitJSON()
		return
	}

	ctx.SetRedirect(redirectURI)

	reqTyp := r.Form.Get(params.responseType)

	data := make(map[string]interface{})
	data["Config"] = s.cfg
	var usr *User
	if r.Method == "POST" {
		username := r.Form.Get(loginParams.username)
		password := r.Form.Get(loginParams.password)

		usr = s.validUser(r, username, password)
	}

	// Case we can't find the user. The user-agent is served  with login template
	// that is used to authenticate the user. All re original query paametersare
	// retained by passing them through a template context variable Action.
	if usr == nil {
		data["Action"] = r.URL.String()
		data["Title"] = "login"

		err = s.view.Render(w, s.cfg.LoginTemplate, data)
		if err != nil {
			s.log.Println(err)
		}
		return
	}

	switch reqTyp {
	case requestType.Code:
		grant := newGrant(s.gen.Generate())
		grant.ExpiresIn = s.cfg.AuthorizationExpire

		grant.Scope = scope
		grant.State = state
		grant.ClientID = client.ID

		usr.Grants = append(usr.Grants, grant)
		err = s.q.SaveModel(usr)
		if err != nil {
			ctx.SetErrorState(errorsKeys.ServerError, "", state)
			ctx.InternalError = err
			break
		}

		client.Grants = append(client.Grants, grant)
		err = s.q.SaveModel(client)
		if err != nil {
			ctx.SetErrorState(errorsKeys.ServerError, "", state)
			ctx.InternalError = err
			break
		}
		ctx.SetData(params.code, grant.Code)
		ctx.SetData(params.state, state)

	case requestType.Token:
		ctx.SetRedirectFragment(true)
		grant := newGrant(s.gen.Generate())
		grant.Type = grantType.Implicit
		grant.Scope = scope
		grant.State = state
		grant.RedirectURL = redirectURI
		grant.ClientID = client.ID
		grant.UserID = usr.ID

		_, err = s.finalizeAccess(&grant, ctx)
		if err != nil {
			ctx.SetError(errorsKeys.ServerError, "")
			ctx.InternalError = err
			break
		}
		if state != "" {
			ctx.SetData(params.state, state)
		}

	default:
		ctx.SetErrorState(errorsKeys.UnsupportedResponseType, "", state)

	}
	_ = ctx.CommitJSON()
}

// Access provide oauth 2.0  access. This support all grant rypes specified by RFC 6976 namely
//	* Authorization code
//	* Implicit
//	* Resource owner password credentials
//	* Client credentials
func (s Server) Access(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(w)
	if r.Method == "GET" {
		if !s.cfg.AllowGetAccess {
			ctx.SetError(errorsKeys.InvalidRequest, "")
			ctx.InternalError = errors.New("request must be POSt")
			_ = ctx.CommitJSON()
			return
		}
	}
	if r.Method != "POST" {
		ctx.SetError(errorsKeys.InvalidRequest, "")
		ctx.InternalError = errors.New("request must be POSt")
		_ = ctx.CommitJSON()
		return
	}
	_ = r.ParseForm()

	accessGrant := r.Form.Get(params.grantType)
	redirectURI := r.Form.Get(params.redirectURL)
	scope := r.Form.Get(params.scope)
	code := r.Form.Get(params.code)

	auth, err := getCLientAuth(r, true)
	if err != nil {
		ctx.SetError(errorsKeys.InvalidClient, "")
		ctx.InternalError = err
		_ = ctx.CommitJSON()
		return
	}

	if s.cfg.AccessAllowed(accessGrant) {
		switch accessGrant {
		case grantType.AuthorizationCode:
			if code == "" {
				ctx.SetError(errorsKeys.InvalidGrant, "")
				break
			}

			client := s.getClient(auth)
			if client == nil {
				break
			}

			grant, err := s.q.GrantByCLient(client, code)
			if err != nil {
				ctx.SetError(errorsKeys.UnauthorizedClient, "")
				ctx.InternalError = err
				break
			}

			if grant.IsExpired() {
				ctx.SetError(errorsKeys.InvalidGrant, "")
				break
			}

			if redirectURI == "" {
				redirectURI = firstURI(client.RedirectURL, s.cfg.RedirSeparator)
			}

			if err = validateURIList(client.RedirectURL, redirectURI, s.cfg.RedirSeparator); err != nil {
				ctx.SetError(errorsKeys.InvalidRequest, "")
				ctx.InternalError = err
				break
			}

			_, err = s.finalizeAccess(grant, ctx)
			if err != nil {
				ctx.SetError(errorsKeys.ServerError, "")
				ctx.InternalError = err
				break
			}
		case grantType.RefreshToken:
			refreshToken := r.Form.Get(params.refreshToken)
			if refreshToken == "" {
				ctx.SetError(errorsKeys.InvalidGrant, "")
				break
			}

			client := s.getClient(auth)
			if client == nil {
				break
			}

			grant, err := s.q.GrantByRefreshToken(refreshToken)
			if err != nil {
				ctx.SetError(errorsKeys.InvalidGrant, "")
				ctx.InternalError = err
				break
			}

			if grant.ClientID != client.ID {
				ctx.SetError(errorsKeys.UnauthorizedClient, "")
				ctx.InternalError = err
				break
			}

			authGrant := &Grant{
				Scope:       scope,
				RedirectURL: grant.RedirectURL,
			}

			if authGrant.Scope == "" {
				authGrant.Scope = grant.Scope
			}

			if extraScopes(grant.Scope, authGrant.Scope) {
				ctx.SetError(errorsKeys.AccessDenied, "")
				break
			}
			_, err = s.finalizeAccess(grant, ctx)
			if err != nil {
				ctx.SetError(errorsKeys.ServerError, "")
				ctx.InternalError = err
				break
			}

		case grantType.Password:
			// handle
			username := r.Form.Get("username")
			password := r.Form.Get("password")

			if username == "" || password == "" {
				ctx.SetError(errorsKeys.InvalidGrant, "")
				break
			}

			usr := s.validUser(r, username, password)
			if usr == nil {
				ctx.SetError(errorsKeys.InvalidGrant, "")
				break
			}

			client := s.getClient(auth)
			if client == nil {
				break
			}

			grant := &Grant{
				Scope:    scope,
				UserID:   usr.ID,
				ClientID: client.ID,
			}
			_, err = s.finalizeAccess(grant, ctx)
			if err != nil {
				ctx.SetError(errorsKeys.ServerError, "")
				ctx.InternalError = err
				break
			}

		case grantType.ClientCredentials:
			// handle
			client := s.getClient(auth)
			if client == nil {
				break
			}

			grant := &Grant{
				Scope:    scope,
				ClientID: client.ID, UserID: client.UserID,
			}
			_, err = s.finalizeAccess(grant, ctx)
			if err != nil {
				ctx.SetError(errorsKeys.ServerError, "")
				ctx.InternalError = err
				break
			}

		case grantType.Assertion:
			assertionTyp := r.Form.Get(params.assertionType)
			assertion := r.Form.Get(params.assertion)

			if assertionTyp == "" || assertion == "" {
				ctx.SetError(errorsKeys.InvalidGrant, "")
				break
			}
			client := s.getClient(auth)
			if client == nil {
				break
			}
			redirectURI = firstURI(client.RedirectURL, s.cfg.RedirSeparator)
			grant := &Grant{
				Scope:       scope,
				RedirectURL: redirectURI,
			}

			_, err = s.finalizeAccess(grant, ctx)
			if err != nil {
				ctx.SetError(errorsKeys.ServerError, "")
				ctx.InternalError = err
				break
			}
		}

	} else {
		ctx.SetError(errorsKeys.UnsupportedGrantType, "")
	}
	_ = ctx.CommitJSON()
}

func (s *Server) getClient(auth interface{}) *Client {
	switch auth.(type) {
	case *basicAuth:
		cAuth := auth.(*basicAuth)
		client, err := s.q.ClientByCode(cAuth.UserName)
		if err != nil {
			return nil
		}

		err = compareHashedString(client.Secret, cAuth.Password)
		if err != nil {
			return nil
		}
		return client
	case *bearerAuth:
		// handle bearer auth
		bAuth := auth.(*bearerAuth)

		access, err := s.q.TokenByCode(bAuth.Code)
		if err != nil {
			s.log.Println(err)
			return nil
		}
		client := &Client{}
		dd := s.q.Where(&Client{ID: access.ClientID}).First(client)
		if dd.Error != nil {
			s.log.Println(dd.Error)
			return nil
		}
		return client
	}
	return nil
}

// finalizeAccess finalizess access request by generating access token and refresh token for the access grant.
// When the access grant is saved to the database, the authorize grant is deleted.
func (s *Server) finalizeAccess(authGrant *Grant, ctx *context) (accessGrant *Grant, err error) {
	accessGrant = &Grant{}
	accessGrant.ClientID = authGrant.ClientID
	accessGrant.UserID = authGrant.UserID
	accessGrant.RedirectURL = authGrant.RedirectURL
	accessGrant.Scope = authGrant.Scope
	accessGrant.State = authGrant.State
	accessGrant.ExpiresIn = s.cfg.AccessExpire

	genAccessToken := Token{
		Code:     s.gen.Generate(),
		ClientID: authGrant.ClientID,
		UserID:   authGrant.UserID,
	}

	if err = s.q.SaveModel(&genAccessToken); err != nil {
		return nil, err
	}

	genRefreshToken := Token{
		Code:     s.gen.Generate(),
		ClientID: authGrant.ClientID,
		UserID:   authGrant.UserID,
	}

	if err = s.q.SaveModel(&genRefreshToken); err != nil {
		return nil, err
	}

	accessGrant.AccessToken = genAccessToken
	accessGrant.RefreshToken = genRefreshToken

	if err = s.q.SaveModel(accessGrant); err != nil {
		return nil, err
	}

	ctx.SetData(params.accessToken, accessGrant.AccessToken.Code)
	ctx.SetData(params.tokenType, s.cfg.TokenType)
	ctx.SetData(params.expiresIn, accessGrant.ExpiresIn)

	if accessGrant.RefreshTokenID != 0 {
		ctx.SetData(params.refreshToken, accessGrant.RefreshToken.Code)
	}
	if accessGrant.Scope != "" {
		ctx.SetData(params.scope, accessGrant.Scope)
	}

	if authGrant.ID != 0 {
		// delete authorization
		if aerr := s.q.DeleteModel(authGrant); aerr != nil {
			//TODO ??
		}

		//		// delete access tokens and refresh tokens
		//		if aerr := s.q.DeleteModel(authGrant.AccessToken); aerr != nil {
		//			//TODO ???
		//		}
		//		if aerr := s.q.DeleteModel(authGrant.RefreshToken); aerr != nil {
		//			//TODO ???
		//		}

	}
	return
}

// Info provide user information using Bearer token. The information served is user email, name and
// avatar_url(this is the link to the user's profile picture a.k.a avatar.
func (s *Server) Info(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(w)

	if err := r.ParseForm(); err != nil {
		ctx.SetError(errorsKeys.InvalidRequest, "")
		ctx.InternalError = err
		_ = ctx.CommitJSON()
		return
	}

	bearer := checkBearerAuth(r)
	if bearer == nil {
		ctx.SetError(errorsKeys.InvalidRequest, "")
		_ = ctx.CommitJSON()
		return
	}

	if bearer.Code == "" {
		ctx.SetError(errorsKeys.InvalidRequest, "")
		_ = ctx.CommitJSON()
		return
	}

	client := s.getClient(bearer)
	if client == nil {
		ctx.SetError(errorsKeys.UnauthorizedClient, "")
		_ = ctx.CommitJSON()
		return
	}

	grant, err := s.q.GrantByBearer(bearer.Code)
	if err != nil {
		ctx.SetError(errorsKeys.InvalidGrant, "")
		_ = ctx.CommitJSON()
		return
	}

	if grant.IsExpired() {
		ctx.SetError(errorsKeys.InvalidGrant, "")
		_ = ctx.CommitJSON()
		return
	}

	user, err := s.q.UserByID(grant.UserID)
	if err != nil {
		ctx.SetError(errorsKeys.InvalidGrant, "")
		_ = ctx.CommitJSON()
		return
	}

	switch grant.Scope {
	case "user":
		ctx.SetData("email", user.Email)
		ctx.SetData("avatar_url", "avatar")
		ctx.SetData("name", user.UserName)
	default:
		ctx.SetError(errorsKeys.InvalidGrant, "")
	}
	_ = ctx.CommitJSON()
}

// Register registers a new user.
func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	data["Config"] = s.cfg
	_ = r.ParseForm()
	if r.Method == "POST" {
		username := r.Form.Get(registerParams.username)
		password := r.Form.Get(registerParams.password)
		confirm := r.Form.Get(registerParams.confirm)
		email := r.Form.Get(registerParams.email)

		if username == "" || password == "" || confirm == "" || email == "" {
			// requested fields should not be embty
			return
		}

		if password != confirm {
			// password should equalconfirm
			return
		}

		hpass, err := hashString(password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			data["status"] = http.StatusInternalServerError
			data["message"] = err.Error()
			rerr := s.view.Render(w, s.cfg.ErrorTemplate, data)
			if rerr != nil {
				s.log.Println(rerr)
			}
			return
		}

		user := &User{
			UserName: username,
			Password: hpass,
			Email:    email,
		}

		err = s.q.CreateUser(user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			data["status"] = http.StatusInternalServerError
			data["message"] = err.Error()
			rerr := s.view.Render(w, s.cfg.ErrorTemplate, data)
			if rerr != nil {
				s.log.Println(rerr)
			}
			return
		}

		http.Redirect(w, r, LoginPath, http.StatusFound)
		return
	}

	err := s.view.Render(w, s.cfg.RegisterTemplate, data)
	if err != nil {
		s.log.Println(err)
	}

}

// Login loges in a user.
func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	data["Title"] = "login"
	data["Config"] = s.cfg
	if r.Method == "POST" {
		_ = r.ParseForm()
		usr := s.loginUser(w, r)
		if usr != nil {
			// create session and redirect to the homepage
			serr := s.SaveToSession(w, r, "UserID", usr.ID)
			if serr != nil {
				s.log.Println(serr)
			}
			http.Redirect(w, r, HomePath, http.StatusFound)
			return
		}

		// loggin failed
		return
	}
	err := s.view.Render(w, s.cfg.LoginTemplate, data)
	if err != nil {
		s.log.Println(err)
	}
}
func (s *Server) loginUser(w http.ResponseWriter, r *http.Request) *User {
	_ = r.ParseForm()
	data := make(map[string]interface{})
	data["Config"] = s.cfg
	data["Title"] = "login"
	if r.Method == "POST" {
		username := r.Form.Get(loginParams.username)
		password := r.Form.Get(loginParams.password)

		if usr := s.validUser(r, username, password); usr != nil {
			return usr
		}
	}
	data["Action"] = r.URL.String()

	err := s.view.Render(w, s.cfg.LoginTemplate, data)
	if err != nil {
		s.log.Println(err)
	}
	return nil
}

func (s *Server) validUser(r *http.Request, username, password string) *User {
	var (
		usr *User
		err error
	)
	if isEmail(username) {
		usr, err = s.q.UserByEmail(username)
	} else {
		usr, err = s.q.UserByUserName(username)
	}
	if err != nil {
		s.log.Println(err)
		return nil
	}
	err = compareHashedString(usr.Password, password)
	if err != nil {
		s.log.Println(err)
		return nil
	}
	return usr
}

// Logout deletes sessions and logs out user.
func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	_ = s.DeleteSession(w, r, s.cfg.SessionName)
	http.Redirect(w, r, HomePath, http.StatusFound)
}

// Client is handler for user clients. To avoid creating spaghetti code, this method
// utilizes query parameters to pack various functionality.
//
// The following are the query parameter names and their details
//
// 	uid  => Is the user ID it is an int64 value.
//	uact => string describing user action. Options are create,delete,refresh,update
//	clID => client ID it is int64 value.
func (s *Server) Client(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	u := q.Get("uid")
	uAction := q.Get("uact")

	_ = r.ParseForm()
	data := make(map[string]interface{})
	data[contextParams.Config] = s.cfg

	userID, err := strconv.Atoi(u)
	if err != nil {
		data[contextParams.Message] = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
		_ = s.view.Render(w, s.cfg.ErrorTemplate, data)
		return
	}

	usr, ok := s.isSession(r)
	if !ok || usr.ID != int64(userID) {
		http.Redirect(w, r, LoginPath, http.StatusFound)
		return
	}

	if r.Method == "POST" {
		clientName := r.Form.Get("client_name")
		clientSecret := r.Form.Get("client_secret")
		switch uAction {
		case "create":
			c := &Client{
				Name: clientName,
				UUID: s.gen.Generate(),
			}
			secret, err := hashString(clientSecret)
			if err != nil {
				data[contextParams.Message] = err.Error()
				w.WriteHeader(http.StatusInternalServerError)
				_ = s.view.Render(w, s.cfg.ErrorTemplate, data)
				return
			}
			c.Secret = secret
			usr.Clients = append(usr.Clients, *c)
			if err = s.q.SaveModel(usr); err != nil {
				data[contextParams.Message] = err.Error()
				w.WriteHeader(http.StatusInternalServerError)
				_ = s.view.Render(w, s.cfg.ErrorTemplate, data)
				return
			}
			http.Redirect(w, r, ClientsPath, http.StatusFound)
			return
		}
	}

	switch uAction {
	case "delete":

		// Deelete the client whose id is specified in the url query paramater clID
		// TODO(gernest): choose a decent name for the query parameter instead of
		// clID
		cID := q.Get("clID")
		clientID, err := strconv.Atoi(cID)
		if err != nil {
			data[contextParams.Message] = err.Error()
			w.WriteHeader(http.StatusInternalServerError)
			_ = s.view.Render(w, s.cfg.ErrorTemplate, data)
			return
		}

		client := &Client{}
		d := s.q.Where(&Client{ID: int64(clientID)}).First(client)
		if d.Error != nil {
			data[contextParams.Message] = d.Error.Error()
			w.WriteHeader(http.StatusInternalServerError)
			_ = s.view.Render(w, s.cfg.ErrorTemplate, data)
			return
		}
		if err = s.q.DeleteModel(client); err != nil {
			data[contextParams.Message] = err.Error()
			w.WriteHeader(http.StatusInternalServerError)
			_ = s.view.Render(w, s.cfg.ErrorTemplate, data)
			return
		}
	}
	err = s.view.Render(w, s.cfg.CLientTemplate, data)
	if err != nil {
		s.log.Println(err)
	}

}

// Home renders hero homepage
func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	data["Config"] = s.cfg
	data["Title"] = "Heroes"

	err := s.view.Render(w, s.cfg.HomeTemplate, data)
	if err != nil {
		s.log.Println(err)
	}
}

// SaveToSession saves the key and value into cookie session.
func (s *Server) SaveToSession(w http.ResponseWriter, r *http.Request, key string, value interface{}) error {
	ss, err := s.store.Get(r, s.cfg.SessionName)
	if err != nil {
		s.log.Println(err)
	}
	ss.Values[key] = value
	return ss.Save(r, w)
}

// isSession returns true if the request is loged in session.
func (s *Server) isSession(r *http.Request) (*User, bool) {
	ss, _ := s.store.Get(r, s.cfg.SessionName)
	if uID, ok := ss.Values["UserID"]; ok {
		userID := uID.(int64)
		usr, err := s.q.UserByID(userID)
		if err != nil {
			return nil, false
		}
		return usr, true
	}
	return nil, false
}

//ServeHTTP serves http request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

//Profile handle user profile.
func (s *Server) Profile(w http.ResponseWriter, r *http.Request) {

}

// DeleteSession deletes cookie session named name.
func (s *Server) DeleteSession(w http.ResponseWriter, r *http.Request, namse string) error {
	ss, _ := s.store.Get(r, namse)
	return s.store.Delete(r, w, ss)
}

// Migrate performs database migrations.
func (s *Server) Migrate() {
	fmt.Print("running migrations...")
	s.q.AutoMigrate(&Token{}, &User{}, &Profile{}, &Session{}, &Client{}, &Grant{})
	fmt.Printf("done \n")
}

// DropAllTables drops all database tables used by hero.
func (s *Server) DropAllTables() {
	models := []interface{}{&User{}, &Profile{}, &Token{}, Grant{}, &Client{}, &Session{}}
	for _, table := range models {
		s.q.DropTableIfExists(table)
	}
}

// Run runs hero webserver.
func (s *Server) Run() {
	host := "http://localhost"
	port := 8090
	if s.cfg.Port != 0 {
		port = s.cfg.Port
	}
	s.log.Printf("starting hero service at  %s:%d \n", host, port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), s))
}

// RunTLS runs hero webserver with https.
func (s *Server) RunTLS(cert, key string) {
	host := "https://localhost"
	port := 443
	if s.cfg.Port != 0 {
		port = s.cfg.Port
	}
	s.log.Printf("starting hero service at  %s:%d \n", host, port)
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", port), cert, key, s))
}

// TestClient creates a user usr and a new client c for usr, this is a helper for testing purpose.
func (s *Server) TestClient(usr *User, c *Client) (*User, *Client) {
	hpas, err := hashString(usr.Password)
	if err != nil {
		panic(err)
	}
	cSec, err := hashString(c.Secret)
	if err != nil {
		panic(err)
	}

	usr.Password = hpas
	c.Secret = cSec

	err = s.q.SaveModel(usr)
	if err != nil {
		panic(err)
	}
	c.UserID = usr.ID
	err = s.q.SaveModel(c)
	if err != nil {
		panic(err)
	}
	return usr, c
}

// GetFlashMessages retrieves flash messages from the session
func (s *Server) GetFlashMessages(r *http.Request, w http.ResponseWriter) FlashMessages {
	ss, _ := s.store.Get(r, s.cfg.SessionName)
	if v, ok := ss.Values[FlashKey]; ok {
		delete(ss.Values, FlashKey)
		_ = ss.Save(r, w)
		return v.(FlashMessages)
	}
	return nil
}

// SaveFlashMessages saves flash messages into session
func (s *Server) SaveFlashMessages(r *http.Request, w http.ResponseWriter, f FlashMessages) error {
	ss, _ := s.store.Get(r, s.cfg.SessionName)
	var flashes FlashMessages
	if v, ok := ss.Values[FlashKey]; ok {
		flashes = v.(FlashMessages)
	}
	ss.Values[FlashKey] = append(flashes, f...)
	err := ss.Save(r, w)
	if err != nil {
		return err
	}
	return nil

}

// SetLogger sets l as the main logger.
func (s *Server) SetLogger(l Logger) {
	s.log = l
}
