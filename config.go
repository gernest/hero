package hero

import (
	"io/ioutil"
	"path/filepath"
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
