package hero

import "time"

// User is hero user object.
type User struct {
	ID        int64
	UserName  string
	Email     string
	Avatar    string
	Profile   Profile
	ProfileID int64
	Grants    []Grant
	Tokens    []Token
	Clients   []Client
	Password  string
	CreatedAt time.Time
}

// Profile is user's profile information
type Profile struct {
	ID        int64
	FirstName string
	LastName  string
	UserName  string
	Email     string
	AvatarURL string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Token is a hero token object.
type Token struct {
	ID        int64
	Code      string
	ClientID  int64
	UserID    int64
	ExpiresIn int64
	CreatedAT time.Time
	UpdatedAt time.Time
}

// Client is a hero client object.
type Client struct {
	ID          int64
	UUID        string
	UserID      int64
	Name        string
	Secret      string
	Grants      []Grant
	Tokens      []Token
	RedirectURL string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Session stores session data from gorilla/sessions
type Session struct {
	ID        int64
	Key       string
	Data      string `sql:"type:text"`
	ExpiresOn time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Grant is a hero grant object.
type Grant struct {
	ID               int64
	Code             string
	Type             string
	UserID           int64
	ClientID         int64
	AccessToken      Token
	AccessTokenID    int64
	AuthorizeToken   Token
	AuthorizeTokenID int64
	RefreshToken     Token
	RefreshTokenID   int64
	Scope            string
	State            string
	RedirectURL      string
	ExpiresIn        int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsExpired returns true if the grant is expired.
func (g *Grant) IsExpired() bool {
	return g.CreatedAt.Add(time.Duration(g.ExpiresIn) * time.Second).Before(time.Now())
}

// TokenGenerator is an interface for generating new tokens.
type TokenGenerator interface {
	Generate() string
}

func newToken(code string) Token {
	return Token{Code: code}
}
func newGrant(code string) Grant {
	return Grant{Code: code}
}
