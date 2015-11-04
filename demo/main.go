// This code is adopted from the examples found in https://github.com/stretchr/gomniauth/example/nethttp

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gernest/hero"
	"github.com/gernest/hero/client"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
)

func main() {
	usr := hero.User{
		UserName: "hero",
		Password: "hero",
		Email:    "hero@swordsplay.com",
	}

	genericClient := hero.Client{
		Name:   "simple",
		UUID:   "sampleUUID",
		Secret: "mysecret",
	}

	heroCfg := hero.DefaultConfig()

	heroURL := "http://localhost:8000"
	demoserver := "http://localhost:8001"

	s := hero.NewServer(heroCfg, &hero.SimpleTokenGen{}, nil)
	s.DropAllTables()
	s.Migrate()
	cCliet := genericClient
	cCliet.RedirectURL = demoserver + "/callback"
	cUsr := usr
	s.TestClient(&cUsr, &cCliet)

	clientCfg := &client.Config{
		ProviderName:        "hero",
		ProviderDisplayName: "Hero",
		AuthURL:             fmt.Sprintf("%s%s", heroURL, heroCfg.AuthEndpoint),
		TokenURL:            fmt.Sprintf("%s%s", heroURL, heroCfg.TokenEndpoint),
		ProfileURL:          heroURL + heroCfg.InfoEndpoint,
		CLientID:            genericClient.UUID,
		CLientSecret:        genericClient.Secret,
		DefaultScope:        "user",
		RedirectURL:         fmt.Sprintf("%s/callback", demoserver),
	}
	gomniauth.SetSecurityKey("ylqRcG4sLnhgOUIt3hbPKiHULHgrutOkpBNwibeJjL4eZ08zzR6YQ0WPl476Cubo")
	gomniauth.WithProviders(
		client.New(clientCfg),
	)
	demo := http.NewServeMux()

	demo.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		provider, err := gomniauth.Provider(clientCfg.ProviderName)
		if err != nil {
			//			w.Write([]byte(err.Error()))
			//			return
			panic(err)
		}
		state := gomniauth.NewState("after", "success")

		// This code borrowed from goweb example and not fixed.
		// if you want to request additional scopes from the provider,
		// pass them as login?scope=scope1,scope2
		//options := objx.MSI("scope", ctx.QueryValue("scope"))

		authUrl, err := provider.GetBeginAuthURL(state, nil)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// redirect
		http.Redirect(w, r, authUrl, http.StatusFound)
	})

	demo.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		provider, err := gomniauth.Provider(clientCfg.ProviderName)
		if err != nil {
			panic(err)
		}
		omap, err := objx.FromURLQuery(r.URL.RawQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		creds, err := provider.CompleteAuth(omap)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		/*
			// This code borrowed from goweb example and not fixed.
			// get the state
			state, err := gomniauth.StateFromParam(ctx.QueryValue("state"))

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// redirect to the 'after' URL
			afterUrl := state.GetStringOrDefault("after", "error?e=No after parameter was set in the state")

		*/

		// load the user
		user, userErr := provider.GetUser(creds)

		if userErr != nil {
			http.Error(w, userErr.Error(), http.StatusInternalServerError)
			return
		}

		rst := make(map[string]interface{})
		rst["name"] = user.Name()
		rst["email"] = user.Email()
		json.NewEncoder(w).Encode(rst)

		// redirect
		//return goweb.Respond.WithRedirect(ctx, afterUrl)
	})

	go http.ListenAndServe(":8000", s)
	log.Println(" visit server at " + demoserver + "/login")
	log.Fatal(http.ListenAndServe(":8001", demo))
}
