package hero

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

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

//ProfileUpdate handle updating user profile.
func (s *Server) ProfileUpdate(w http.ResponseWriter, r *http.Request) {

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
