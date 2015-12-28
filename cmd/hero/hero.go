package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/codegangsta/cli"
	"github.com/gernest/hero"
	"github.com/koding/multiconfig"
)

var version = "0.0.1"
var (
	defaultCfg = hero.DefaultConfig()
)

const (
	configName = "config.json"
)

func authors() []cli.Author {
	return []cli.Author{
		cli.Author{
			Name:  "Geofrey Ernest",
			Email: "geofreyernest@live.com",
		},
	}
}

func getConfig(path string) (*hero.Config, error) {
	if path == "" {
		path = configName
	}
	loader := multiconfig.MultiLoader(
		&multiconfig.TagLoader{},
		&multiconfig.EnvironmentLoader{},
		&multiconfig.JSONLoader{Path: path},
	)
	d := &multiconfig.DefaultLoader{}
	d.Loader = loader
	d.Validator = multiconfig.MultiValidator(&multiconfig.RequiredValidator{})

	cfg := &hero.Config{}

	err := d.Load(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func serverCommand() cli.Command {
	return cli.Command{
		Name:      "server",
		ShortName: "s",
		Usage:     "runs hero aouth 2 service",
		Action:    server,
		Flags: []cli.Flag{
			cli.BoolTFlag{
				Name:  "migrate",
				Usage: "creates database tables if they don't exist",
			},
			cli.BoolTFlag{
				Name:  "dev",
				Usage: "enable development server",
			},
			cli.BoolFlag{
				Name:  "https",
				Usage: "enable https",
			},
			cli.StringFlag{
				Name:   "cert",
				Usage:  "The filepath to a ssl certificate",
				EnvVar: "HERO_HTTPS_CERT",
			},
			cli.StringFlag{
				Name:   "key",
				Usage:  "The filepath to a rsa key used to sign the ssl certificate",
				EnvVar: "HERO_HTTPS_KEY",
			},
		},
	}
}

func server(ctx *cli.Context) {
	cfgFile := configName
	if first := ctx.Args().First(); first != "" {
		cfgFile = first
	}

	cfg, err := getConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	s := hero.NewServer(cfg, &hero.SimpleTokenGen{}, nil)

	if ctx.BoolT("dev") {
		v, err := hero.NewDefaultView(cfg.TemplatesDir, true)
		if err != nil {
			fmt.Println(err)
			return
		}
		s = hero.NewServer(cfg, &hero.SimpleTokenGen{}, v)
	}
	if ctx.BoolT("migrate") {
		s.Migrate()
	}
	if ctx.BoolT("https") {
		cert := ctx.String("cert")
		_, ferr := os.Stat(cert)
		if ferr != nil {
			fmt.Println(ferr)
			return
		}
		key := ctx.String("key")
		_, ferr = os.Stat(cert)
		if ferr != nil {
			fmt.Println(ferr)
			return
		}
		s.RunTLS(cert, key)
		return
	}

	s.Run()
}

func generateCommand() cli.Command {
	return cli.Command{
		Name:      "genconf",
		ShortName: "g",
		Usage:     "generate default configurations",
		Action:    genconfig,
	}
}

func writeConfig(cfg *hero.Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0600)
}

func genconfig(ctx *cli.Context) {
	cfgFile := configName
	if arg := ctx.Args().First(); arg != "" {
		cfgFile = arg
	}
	err := writeConfig(defaultCfg, cfgFile)
	if err != nil {
		fmt.Println(err)
	}

}

func main() {
	app := cli.NewApp()
	app.Name = "hero"
	app.Version = version
	app.Usage = "Oauth2 provider"
	app.Commands = []cli.Command{
		serverCommand(),
		generateCommand(),
	}
	app.Run(os.Args)
}
