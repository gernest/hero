## Configuration

This is content of an example default configuration file. You can also find it [fixtures/config.json](fixtures/config.json)

```json
{
	"redirect_separator": "",
	"authorization_expire": 200,
	"access_expire": 200,
	"allow_get_access": false,
	"allowed_access_type": [
		"authorization_code",
		"refresh_token",
		"password",
		"client_credentials",
		"assertion"
	],
	"token_type": "",
	"provider_name": "",
	"auth_endpoint": "/authorize",
	"token_endpoint": "/tokens",
	"info_endpoint": "/info",
	"port": 8090,
	"database_dialect": "postgres",
	"database_connection": "postgres://postgres:postgres@localhost/hero_test?sslmode=disable",
	"templates_dir": "views",
	"static_dir": "",
	"session_path": "/",
	"session_max_age": 0,
	"session_domain": "",
	"session_secure": false,
	"session_hhhponly": false,
	"session_name": "_hero",
	"Login_template": "login.html",
	"error_template": "error.html",
	"register_template": "",
	"client_template": "client.html",
	"profile_template": "profile.html",
	"home_template": "home.html"
}
```

Table to explain the configuration settings

setting               | type      | details
----------------------|-----------|----------------------
redirect_separator    |  string   | character used to separate multiple redirect urls e.g `:`
authorization_expire  |  int64    | duration in seconds of the authorization code
access_expire         |  int64    | duration in seconds of the access code
allow_get_access      |  bool     | if true allow GET requests
AllowedAccess_type    |  []string | allowed access types e.g `["refresh_token","password"]`
token_type            |  string   | the type of tokens
provider_name         |  string   | the name of the provider( your server) e.g hero
auth_endpoint         |  string   | the route where authorization is handeld e.g /auhorize
token_endpoint        |  string   | the route where access is handled e.g /tokens
info_endpoint         |  string   | the route where the granted user details are accessed.
port                  |  int      | port number where the server will be listenig to. e.g 8080
database_dialect      |  string   | type of database e.g postgres, mysql, foundation
database+connection   |  string   | database connection url
templates_dir         |  string   | the directory where templates are stored.
static_dir            |  string   | the directory where static assets are stored i.e javascript, stylesheets  etc
session_path          |  string   | path of the session( cookie)
session_max_age       |  int      | duration of the cookie(session)
session_domain        |  string   | deomain name for the session
session_secure        |  bool     | true if the sessio is secure( https)
session_httponly      |  bool     | true if session is hhtp only
session_name          |  string   | the name of the session e.g _a_mist_of_avalon
login_template        |  string   | the name of the template to render for login users
error_template        |  string   | the name of the template to render on errors
register_template     |  string   | the name of the template to render on registering users
cLient_template       |  string   | the name of template to render on create/read/update/delete clients
profile_template      |  string   | the name of the template to render on user profile
home_template         |  string   | the name of the template to render at home page

