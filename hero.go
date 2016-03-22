// Package hero is a heroic oauth2 provider.
package hero

import (
	"errors"
	"net/http"
	"net/url"

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

	//ProfileUpdatePath is the route for updating user profile
	ProfileUpdatePath = "/profile/update"

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
	s.mux.HandleFunc(ProfileUpdatePath, s.ProfileUpdate).Methods("GET", "POST")
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
