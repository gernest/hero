package hero

import (
	"encoding/base32"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
)

const (
	defaultSessionMaxAge = 2592000
	defaultSessionPath   = "/"
)

// Store is a Store implementation for gorilla session. It support mysql, postgresql and
// foundation databases
type Store struct {
	q       *query
	codecs  []securecookie.Codec
	options *sessions.Options
}

// DefaultStore returns a *Store with default values.
func DefaultStore(db *gorm.DB) *Store {
	keyPairs := [][]byte{
		[]byte("ePAPW9vJv7gHoftvQTyNj5VkWB52mlza"),
		[]byte("N8SmpJ00aSpepNrKoyYxmAJhwVuKEWZD"),
	}
	cfg := &Config{
		SessionMaxAge: defaultSessionMaxAge,
		SessionPath:   defaultSessionPath,
	}
	return NewStore(db, cfg, keyPairs...)
}

// NewStore creates a new *Store instance.
func NewStore(db *gorm.DB, config *Config, keyPairs ...[]byte) *Store {
	q := &query{}
	q.DB = db
	return &Store{
		q:      q,
		codecs: securecookie.CodecsFromPairs(keyPairs...),
		options: &sessions.Options{
			Path:     config.SessionPath,
			Domain:   config.SessionDomain,
			MaxAge:   config.SessionMaxAge,
			Secure:   config.SessionSecure,
			HttpOnly: config.SessionHTTPOnly,
		},
	}
}

// Get fetches a session for a given name after it has been added to the registry.
func (s *Store) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a new session
func (s *Store) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.options
	session.Options = &(opts)
	session.IsNew = true

	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.codecs...)
		if err == nil {
			err = s.load(session)
			if err == nil {
				session.IsNew = false
			}
		}
	}
	return session, err
}

// Save saves the session into a postgresql database
func (s *Store) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// Set delete if max-age is < 0
	if session.Options.MaxAge < 0 {
		if err := s.Delete(r, w, session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		// Generate a random session ID key suitable for storage in the DB
		session.ID = strings.TrimRight(
			base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32)), "=")
	}

	if err := s.save(session); err != nil {
		return err
	}

	// Keep the session ID key in a cookie so it can be looked up in DB later.
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.codecs...)
	if err != nil {
		return err
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

//load fetches a session by ID from the database and decodes its content into session.Values
func (s *Store) load(session *sessions.Session) error {
	ss, err := s.q.GetSessionByKey(session.ID)
	if err != nil {
		return err
	}
	return securecookie.DecodeMulti(session.Name(), string(ss.Data),
		&session.Values, s.codecs...)
}

func (s *Store) save(session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values,
		s.codecs...)

	if err != nil {
		return err
	}

	var expiresOn time.Time

	exOn := session.Values["expires_on"]

	if exOn == nil {
		expiresOn = time.Now().Add(time.Second * time.Duration(session.Options.MaxAge))
	} else {
		expiresOn = exOn.(time.Time)
		if expiresOn.Sub(time.Now().Add(time.Second*time.Duration(session.Options.MaxAge))) < 0 {
			expiresOn = time.Now().Add(time.Second * time.Duration(session.Options.MaxAge))
		}
	}
	ss := &Session{
		Key:       session.ID,
		Data:      encoded,
		ExpiresOn: expiresOn,
	}
	if session.IsNew {
		return s.q.SaveSession(ss)
	}
	return s.q.UpdateSession(ss)
}

func (s *Store) destroy(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	options := *s.options
	options.MaxAge = -1
	http.SetCookie(w, sessions.NewCookie(session.Name(), "", &options))
	for k := range session.Values {
		delete(session.Values, k)
	}
	return s.q.DeleteSession(session.ID)
}

// Delete deletes session.
func (s *Store) Delete(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return s.destroy(r, w, session)
}
