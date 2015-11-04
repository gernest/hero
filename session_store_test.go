package hero

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

type headerOnlyResponseWriter http.Header

func (ho headerOnlyResponseWriter) Header() http.Header {
	return http.Header(ho)
}

func (ho headerOnlyResponseWriter) Write([]byte) (int, error) {
	panic("NOIMPL")
}

func (ho headerOnlyResponseWriter) WriteHeader(int) {
	panic("NOIMPL")
}

var secret = "EyaC2BPcJtNqU3tjEHy+c+Wmqc1yihYIbUWEl/jk0Ga73kWBclmuSFd9HuJKwJw/Wdsh1XnjY2Bw1HBVph6WOw=="

func TestStore(t *testing.T) {
	if !dbConn.isOpne {
		t.Skip()
	}

	ss := DefaultStore(dbConn.db)

	if ss == nil {
		t.Fatal("This test requires a real database")
	}

	// ROUND 1 - Check that the cookie is being saved
	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("failed to create request", err)
	}

	session, err := ss.Get(req, "mysess")
	if err != nil {
		t.Fatal("failed to get session", err.Error())
	}

	session.Values["counter"] = 1

	m := make(http.Header)
	if err = ss.Save(req, headerOnlyResponseWriter(m), session); err != nil {
		t.Fatal("Failed to save session:", err.Error())
	}

	if m["Set-Cookie"][0][0:6] != "mysess" {
		t.Fatal("Cookie wasn't set!")
	}

	// ROUND 2 - check that the cookie can be retrieved
	req, err = http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("failed to create round 2 request", err)
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, ss.codecs...)
	if err != nil {
		t.Fatal("Failed to make cookie value", err)
	}

	req.AddCookie(sessions.NewCookie(session.Name(), encoded, session.Options))

	session, err = ss.Get(req, "mysess")
	if err != nil {
		t.Fatal("failed to get round 2 session", err.Error())
	}

	if session.Values["counter"] != 1 {
		t.Fatal("Retrieved session had wrong value:", session.Values["counter"])
	}

	session.Values["counter"] = 9 // set new value for round 3
	if err = ss.Save(req, headerOnlyResponseWriter(m), session); err != nil {
		t.Fatal("Failed to save session:", err.Error(), *session)
	}

	// ROUND 3 - check that the cookie has been updated
	req, err = http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("failed to create round 3 request", err)
	}
	req.AddCookie(sessions.NewCookie(session.Name(), encoded, session.Options))

	session, err = ss.Get(req, "mysess")
	if err != nil {
		t.Fatal("failed to get session round 3", err.Error())
	}

	if session.Values["counter"] != 9 {
		t.Fatal("Retrieved session had wrong value in round 3:", session.Values["counter"])
	}

	// RIUND 4 - check that the cookie can be deleted.

	w := httptest.NewRecorder()
	session.Options.MaxAge = -1

	err = ss.Save(req, w, session)
	if err != nil {
		t.Errorf("saving session %v", err)
	}

	session, err = ss.Get(req, "mysess")
	if err != nil {
		t.Fatal("failed to get session round 4", err.Error())
	}
	c := session.Values["counter"]
	if c != nil {
		t.Errorf("expected nil got %v", c)
	}

}

func TestSessionOptionsAreUniquePerSession(t *testing.T) {
	if !dbConn.isOpne {
		t.Skip()
	}

	ss := DefaultStore(dbConn.db)

	if ss == nil {
		t.Fatal("This test requires a real database")
	}

	ss.options.MaxAge = 900

	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("Failed to create request", err)
	}

	session, err := ss.Get(req, "newsess")
	if err != nil {
		t.Fatal("Failed to create session", err)
	}

	session.Options.MaxAge = -1

	if ss.options.MaxAge != 900 {
		t.Fatalf("PGStore.Options.MaxAge: expected %d, got %d", 900, ss.options.MaxAge)
	}
}
