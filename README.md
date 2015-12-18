# hero [![Build Status](https://travis-ci.org/gernest/hero.svg)](https://travis-ci.org/gernest/hero)  [![GoDoc](https://godoc.org/github.com/gernest/hero?status.svg)](https://godoc.org/github.com/gernest/hero)
hero is a heroic oauth 2 provider, which is simple, elegant and beautiful.


# What is oauth 2 ?

According to [this post](https://www.digitalocean.com/community/tutorials/an-introduction-to-oauth-2)

>OAuth 2 is an authorization framework that enables applications to obtain limited access to user accounts on >an HTTP service, such as Facebook, GitHub, and DigitalOcean.

# What is hero?

From the definition of oauth 2 above, there is example of providers that is Facebook,GitHub and DigitalOcean. `hero` enables you to become an oauth 2 provider just like Facebook ,GitHub and DigitalOcean.

This means, hero provides user accounts and also allows  limited access to the accounts.Unlike the other options, with hero you have full control of everything.


Hero is a commadline application.  It also offers a library that you can use to compose your own version of oauth 2 provider.

# Features

* User account management
* Encrypted passwords and client secrets.
* Clients management
* Client authorization
* Client authentication
* oauth 2 [rfc 6974](http://tools.ietf.org/html/rfc6749) compliant
* User profiles
* Configurable.
* Multiple databases( postgres, mysql and foundation)
* Hot templates reload for rapid development(no need to recompile when run with `-dev` flag).


# Project Status
Hero is in early development a.k.a alpha state.


# Motivation
There is no oauth provider in my country(Tanzania), so I created one in a way that anyone who wants to be can have a starting point.


# Installation

## From source

Prerequisite
* go(Golang) 1.5+

Go get the project to install it

	go get github.com/gernest/hero/cmd/hero

This will create the hero binary for you.

## Precompiled
Alternatively you can download precompiled binaries for your favorite operating system.

[COMING SOON]

# Usage
You need a configuration file  in order to run `hero` server. The format supported is json. Hero comes with a command to bootstrap a configuration file with default settings, you can use it to customize the values as you fancy.I will explain the configurable details in a moment.


# Step 1 Configure

To generate default configuration file, run the following command,

	hero genconf [Path to configuration file goes here e.g config.json]


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


Making a configuarble application is hard, so take this with a grain of salt.


### A note about views a.k.a templates used by `hero`

The names of the templates needed by hero are found in the configuration file. If you are using the `hero` commandline application make sure the names are relative to the `templates_dir` configuration value.

You can create your own views, hero uses [text/template]() templates. You should Invest in differnetiating your brand. hero ships with ready to use, ready to customize views for you, they can be found here [views](views). You can help improve the default views, hero welcomes your contribution.

The following are the templates names and the form fields that are expected to be present when submitting values to the server.

The `template name` part of the tample refers to the value supplied in the configuration file.

template name    | form value| data passed to template
-----------------|-----------|--------------------------
home_template    | none      | `.Config`
login_template   |`login_username`, `login_password`| `.Config`
register_template|`register_username`,`register_password`, `register_confirm`,`register_email`|`.Config`
client_template  | `client_name`, `client_secret`| `.Config`
profile_template | none | `.Config`
error_template   | none| `.Config`,`.Message`,`.StatusCode`

The following is the table of data(context) passed to templates and their meaning

 context data| details
-------------|---------------------------------------------------------
`.Config`    | The configuration object of the server.
`.User`      | The instance of a loged in user profile.
`.Message`   | The error message passed to the `error_template`
`.StatusCode`| http status code passed yo the `error_template` . This is to help customize error rendering.


# Step 2: Run


Run the service

	hero --mingrate server [config_file_path]

where `config_file_path` is the path to the configuration file, it doisnt matter if the path is relative or absolute.

It is wise to add the `--migrate` flag if you are running for the first time so as to prepare the database.


You should see this on your `stdout`

	running migrations...done 
	starting hero service at  http://localhost:8090 


If you see anything diffenernt then open an issue.

# Contributing
contributions are welcome. For development please see [DEVELOPER.md](DEVELOPER.md)


# Author
Geofrey Ernest

Twitter  : [@gernesti](https://twitter.com/gernesti)



# Aknowledgement

The project [osin](https://github.com/RangelReale/osin) is the foundation from which I got insight and direction of how oauth 2 works.



# TODO

*  ~~Fix a lot of typos (English is my third language).~~
*  implement the views templates( I need help here, I suck at styles)
*  Support https.
*  Add extensive documentation.
*  ~~Demo with examples.~~
*  Let's encrypt integration.
*  ~~Add logging support.~~
*  ~~Templates hot reload.~~
*  Livereload.
*  XSRF protection.
*  ~~Flash messages.~~

# Licence

This project is released under the MIT licence. See [LICENCE](LICENCE) for more details.
