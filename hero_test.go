package hero

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/antonholmquist/jason"
)

const formURLEncoded = "application/x-www-form-urlencoded"

var genericUser = User{
	UserName: "gernest",
	Password: "hero",
	Email:    "hero@swordsplay.com",
}

var genericClient = Client{
	Name:   "simple",
	UUID:   "sampleUUID",
	Secret: "mysecret",
}

var testCode string

func TestServer_Home(t *testing.T) {
	req, _ := http.NewRequest("GET", HomePath, nil)
	w := httptest.NewRecorder()
	testServer.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected  %d got %d", http.StatusFound, w.Code)
	}
}

func TestServer_Register(t *testing.T) {
	if !dbConn.isOpne {
		t.Skip()
	}

	regVars := url.Values{
		registerParams.username: {genericUser.UserName},
		registerParams.email:    {genericUser.Email},
		registerParams.password: {genericUser.Password},
		registerParams.confirm:  {genericUser.Password},
	}

	req, err := http.NewRequest("GET", RegisterPath, nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	testServer.ServeHTTP(w, req)
	doc, err := goquery.NewDocumentFromReader(w.Body)
	if err != nil {
		t.Error(err)
	}
	tokField := doc.Find("input").First().Get(0)
	var tok string
	for _, v := range tokField.Attr {
		if v.Key == "value" {
			tok = v.Val
		}
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected  %d got %d", http.StatusFound, w.Code)
	}
	regVars.Set("gorilla.csrf.Token", tok)

	cookies := readSetCookies(w.HeaderMap)
	req, err = http.NewRequest("POST", RegisterPath, strings.NewReader(regVars.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)
	for _, v := range cookies {
		req.AddCookie(v)
	}
	w = httptest.NewRecorder()
	testServer.ServeHTTP(w, req)
	if w.Code != http.StatusFound {
		t.Errorf("expected %d got %d", http.StatusFound, w.Code)
	}
}

func TestServer_Login(t *testing.T) {
	if !dbConn.isOpne {
		t.Skip()
	}

	req, _ := http.NewRequest("GET", LoginPath, nil)
	w := httptest.NewRecorder()
	testServer.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected  %d got %d", http.StatusFound, w.Code)
	}
	doc, err := goquery.NewDocumentFromReader(w.Body)
	if err != nil {
		t.Error(err)
	}
	tokField := doc.Find("input").First().Get(0)
	var tok string
	for _, v := range tokField.Attr {
		if v.Key == "value" {
			tok = v.Val
		}
	}

	logVars := url.Values{
		loginParams.username: {genericUser.UserName},
		loginParams.password: {genericUser.Password},
	}
	logVars.Set("gorilla.csrf.Token", tok)
	cookies := readSetCookies(w.HeaderMap)

	req, err = http.NewRequest("POST", LoginPath, strings.NewReader(logVars.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)
	for _, v := range cookies {
		req.AddCookie(v)
	}
	w = httptest.NewRecorder()
	testServer.ServeHTTP(w, req)
	if w.Code != http.StatusFound {
		t.Errorf("expected %d got %d", http.StatusFound, w.Code)
	}
}

func TestServer_Logout(t *testing.T) {
	if !dbConn.isOpne {
		t.Skip()
	}
	req, _ := http.NewRequest("GET", LogoutPath, nil)
	w := httptest.NewRecorder()

	user, err := testServer.q.UserByEmail(genericUser.Email)
	if err != nil {
		t.Fatal(err)
	}
	testServer.SaveToSession(w, req, "UserID", user.ID)

	if _, ok := testServer.isSession(req); !ok {
		t.Error("expcted session")
	}
	testServer.ServeHTTP(w, req)
	if w.Code != http.StatusFound {
		t.Errorf("expected %d got %d", http.StatusFound, w.Code)
	}
	if _, ok := testServer.isSession(req); ok {
		t.Error("expcted session")
	}
}

func TestServer_Authorize(t *testing.T) {
	if !dbConn.isOpne {
		t.Skip()
	}

	// create a new client that belongs to the genericUser
	user, err := testServer.q.UserByEmail(genericUser.Email)
	if err != nil {
		t.Fatal(err)
	}
	secureSecret, err := hashString(genericClient.Secret)
	if err != nil {
		t.Error(err)
	}
	client := Client{
		Name:   genericClient.Name,
		UUID:   genericClient.UUID,
		Secret: secureSecret,
	}
	user.Clients = append(user.Clients, client)
	err = testServer.q.SaveModel(user)
	if err != nil {
		t.Error(err)
	}

	// store the client UUUID to the global genericClient, to be used in the
	// subsequest tests
	iC := &genericClient
	iC.UUID = client.UUID
	authPath := testServer.cfg.AuthEndpoint

	//
	// case no any form values
	//
	authParams := url.Values{}
	req, err := http.NewRequest("POST", authPath, strings.NewReader(authParams.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)
	w := httptest.NewRecorder()
	testServer.ServeHTTP(w, req)
	jObj, err := jason.NewObjectFromReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	// check error key
	resErr, err := jObj.GetString("error")
	if err != nil {
		t.Error(err)
	}
	if resErr != errorsKeys.ServerError {
		t.Errorf("expected %s got %s", errorsKeys.ServerError, resErr)
	}

	// check the error description, it should return error description for
	// errorKeys.ServerError
	resDescription, err := jObj.GetString("error_description")
	if err != nil {
		t.Error(err)
	}
	errDescription := baseOauthErrs.Get(errorsKeys.ServerError)
	if resDescription != errDescription {
		t.Errorf("expected %s got %s", errDescription, resDescription)
	}

	//
	// case only the client_id, the client has no RedirectURL
	//
	authParams = url.Values{
		params.clientID: {client.UUID},
	}

	req, err = http.NewRequest("POST", authPath, strings.NewReader(authParams.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)
	w = httptest.NewRecorder()
	testServer.ServeHTTP(w, req)
	jObj, err = jason.NewObjectFromReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	// check error key
	resErr, err = jObj.GetString("error")
	if err != nil {
		t.Error(err)
	}
	if resErr != errorsKeys.UnauthorizedClient {
		t.Errorf("expected %s got %s", errorsKeys.UnauthorizedClient, resErr)
	}

	// check the error description, it should return error description for
	// errorKeys.UnauthoredClient
	resDescription, err = jObj.GetString("error_description")
	if err != nil {
		t.Error(err)
	}
	errDescription = baseOauthErrs.Get(errorsKeys.UnauthorizedClient)
	if resDescription != errDescription {
		t.Errorf("expected %s got %s", errDescription, resDescription)
	}

	//
	// case cleint_id with RedirectURL
	//
	c, _ := testServer.q.ClientByCode(client.UUID)
	c.RedirectURL = "http://example.com"
	err = testServer.q.SaveModel(c)
	if err != nil {
		t.Error(err)
	}
	req, err = http.NewRequest("POST", authPath, strings.NewReader(authParams.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)
	w = httptest.NewRecorder()
	testServer.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), "login") {
		t.Error("should render the login view")
	}

	//
	// case request_type is code
	//
	authParams = url.Values{
		params.clientID:     {client.UUID},
		params.responseType: {requestType.Code},
	}
	req, err = http.NewRequest("POST", authPath, strings.NewReader(authParams.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)
	w = httptest.NewRecorder()
	testServer.ServeHTTP(w, req)

	//
	// case client_id and client_secret
	//
	authParams.Set(loginParams.username, genericUser.UserName)
	authParams.Set(loginParams.password, genericUser.Password)
	req, err = http.NewRequest("POST", authPath, strings.NewReader(authParams.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)
	w = httptest.NewRecorder()
	testServer.ServeHTTP(w, req)
	if w.Code != http.StatusFound {
		t.Errorf("expected %d got %d", http.StatusFound, w.Code)
	}
	q, err := url.ParseRequestURI(w.Header().Get("Location"))
	if err != nil {
		t.Error(err)
	}
	// get the code that is returned from the redirect url
	code := q.Query().Get("code")
	if code == "" {
		t.Error("expected grant code")
	}
	testCode = code

	tokenParams := url.Values{
		params.clientID:      {client.UUID},
		params.clientSecret:  {client.Secret},
		params.responseType:  {requestType.Token},
		loginParams.username: {genericUser.UserName},
		loginParams.password: {genericUser.Password},
	}

	req, err = http.NewRequest("POST", authPath, strings.NewReader(tokenParams.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)
	w = httptest.NewRecorder()
	testServer.ServeHTTP(w, req)
	if w.Code != http.StatusFound {
		t.Errorf("expected %d got %d", http.StatusFound, w.Code)
	}
}

func TestServer_Access(t *testing.T) {
	if !dbConn.isOpne {
		t.Skip()
	}

	// check AllowGetAccess
	req, err := http.NewRequest("GET", testServer.cfg.TokenEndpoint, nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	testServer.ServeHTTP(w, req)

	//
	// case unsupported http METHOD
	//
	req, err = http.NewRequest("PATCH", testServer.cfg.TokenEndpoint, nil)
	if err != nil {
		t.Error(err)
	}
	w = httptest.NewRecorder()

	testServer.ServeHTTP(w, req)

	//
	// case no form values
	//
	accessParams := url.Values{}
	req, err = http.NewRequest("POST", testServer.cfg.TokenEndpoint, strings.NewReader(accessParams.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)

	w = httptest.NewRecorder()

	testServer.ServeHTTP(w, req)

	//
	// case request_type authorization_code
	// and the code is not set
	//
	accessParams = url.Values{
		params.clientID:     {genericClient.UUID},
		params.clientSecret: {genericClient.Secret},
		params.grantType:    {grantType.AuthorizationCode},
	}
	req, err = http.NewRequest("POST", testServer.cfg.TokenEndpoint, strings.NewReader(accessParams.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)
	w = httptest.NewRecorder()
	testServer.ServeHTTP(w, req)

	// case a bad code
	accessParams.Set(params.code, "bad xodw")
	req, err = http.NewRequest("POST", testServer.cfg.TokenEndpoint, strings.NewReader(accessParams.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)
	w = httptest.NewRecorder()
	testServer.ServeHTTP(w, req)
	jObj, err := jason.NewObjectFromReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	// check error key
	resErr, err := jObj.GetString("error")
	if err != nil {
		t.Error(err)
	}
	if resErr != errorsKeys.UnauthorizedClient {
		t.Errorf("expected %s got %s", errorsKeys.UnauthorizedClient, resErr)
	}

	// case a good code
	accessParams.Set(params.code, testCode)
	req, err = http.NewRequest("POST", testServer.cfg.TokenEndpoint, strings.NewReader(accessParams.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)
	w = httptest.NewRecorder()
	testServer.ServeHTTP(w, req)

	//
	// Access using refesh token
	//
	accessParams.Set(params.grantType, grantType.RefreshToken)
	jObj, err = jason.NewObjectFromReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	// checkrefresh token
	refreshTok, err := jObj.GetString("refresh_token")
	if err != nil {
		t.Error(err)
	}
	accessParams.Set(params.refreshToken, refreshTok)
	req, err = http.NewRequest("POST", testServer.cfg.TokenEndpoint, strings.NewReader(accessParams.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)
	w = httptest.NewRecorder()
	testServer.ServeHTTP(w, req)

	//
	// Access using client credentials
	//
	accessParams.Set(params.grantType, grantType.ClientCredentials)
	req, err = http.NewRequest("POST", testServer.cfg.TokenEndpoint, strings.NewReader(accessParams.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", formURLEncoded)
	w = httptest.NewRecorder()
	testServer.ServeHTTP(w, req)
}
