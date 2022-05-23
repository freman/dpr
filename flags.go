package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

type config struct {
	listen       string
	tmpWhitelist string
	username     string
	password     string
	whitelist    networks
	roundcube    string
}

var (
	defaultListen      = ":8111"
	defaultOXWhitelist = "127.0.0.0/8"
	defaultOXUsername  = "dovecot"
	defaultOXPassword  = "dovecot"
	defaultRoundcube   = "https://mail.example.com"
)

func envString(name, def string) string {
	if tmp, ok := os.LookupEnv(name); ok {
		return tmp
	}
	return def
}

func configure() config {

	defaultRoundcube = envString("ROUNDCUBE", defaultRoundcube)

	var cfg config

	flag.StringVar(&cfg.listen, "listen", envString("LISTEN", defaultListen), "Define listening configuration.\nDefaults to '"+defaultListen+"' or the value of the LISTEN env var, if it is set")
	flag.StringVar(&cfg.tmpWhitelist, "ox-whitelist", envString("OX_WHITELIST", defaultOXWhitelist), "A comma seperated whitelist of hosts/networks to expect OX notifications from.\nDefaults to '"+defaultOXWhitelist+"' or the value of the OX_WHITELIST env var, if it is set")
	flag.StringVar(&cfg.username, "ox-username", envString("OX_USERNAME", defaultOXUsername), "Username tor require Dovecot to authenticate with when sending an OX notification.\nDefaults to '"+defaultOXUsername+"' or the value of the OX_USERNAME env var, if it is set")
	flag.StringVar(&cfg.password, "ox-password", envString("OX_PASSWORD", defaultOXPassword), "Password tor require Dovecot to authenticate with when sending an OX notification.\nDefaults to '"+defaultOXPassword+"' or the value of the OX_PASSWORD env var, if it is set")
	flag.StringVar(&cfg.roundcube, "roundcube", envString("ROUNDCUBE", defaultRoundcube), "BaseURL for your Roundcube installation.\nDefaults to '"+defaultRoundcube+"' or the value of the ROUNDCUBE env var, if it is set")

	flag.Parse()

	for _, v := range strings.Split(cfg.tmpWhitelist, ",") {
		v = strings.TrimSpace(v)
		if !strings.Contains(v, "/") {
			v += "/32"
		}

		_, n, err := net.ParseCIDR(v)
		if err != nil {
			fmt.Printf("Unable to parse network %s: %w\n", v, err)
			flag.Usage()
			os.Exit(1)
		}

		cfg.whitelist = append(cfg.whitelist, n)

	}

	return cfg
}
