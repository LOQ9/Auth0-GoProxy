package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/LOQ9/Auth0-GoProxy/proxy"
	"github.com/gorilla/context"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Auth0 Proxy"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:   "p, port",
			Value:  80,
			EnvVar: "AP_PORT",
		},
		cli.StringFlag{
			Name:   "proxy-host",
			EnvVar: "AP_PROXY_HOST",
		},
		cli.IntFlag{
			Name:   "proxy-port",
			Value:  80,
			EnvVar: "AP_PROXY_PORT",
		},
		cli.StringFlag{
			Name:   "proxy-scheme",
			Value:  "http",
			EnvVar: "AP_PROXY_SCHEME",
		},
		cli.StringFlag{
			Name:   "auth0-domain",
			EnvVar: "AP_AUTH0_DOMAIN",
		},
		cli.StringFlag{
			Name:   "auth0-client-id",
			EnvVar: "AP_AUTH0_CLIENT_ID",
		},
		cli.StringFlag{
			Name:   "auth0-client-secret",
			EnvVar: "AP_AUTH0_CLIENT_SECRET",
		},
		cli.StringFlag{
			Name:   "auth0-redirect-uri",
			EnvVar: "AP_AUTH0_REDIRECT_URI",
		},
		cli.StringFlag{
			Name:   "session-secret",
			Value:  "some-secret-key",
			EnvVar: "AP_SESSION_SECRET",
		},
		cli.DurationFlag{
			Name:   "session-timeout",
			Value:  time.Hour * 1,
			EnvVar: "AP_SESSION_TIMEOUT",
		},
	}

	app.Action = func(c *cli.Context) error {
		reverseProxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Host:   fmt.Sprintf("%s:%d", c.String("proxy-host"), c.Int("proxy-port")),
			Scheme: c.String("proxy-scheme"),
		})

		auth0Proxy := proxy.NewAuth0Proxy(proxy.Config{
			ReverseProxy:   reverseProxy,
			Domain:         c.String("auth0-domain"),
			ClientID:       c.String("auth0-client-id"),
			ClientSecret:   c.String("auth0-client-secret"),
			RedirectURI:    c.String("auth0-redirect-uri"),
			SessionSecret:  []byte(c.String("session-secret")),
			SessionTimeout: c.Duration("session-timeout")})

		addr := fmt.Sprintf(":%d", c.Int("port"))
		fmt.Printf("Listening on %s\n", addr)
		return http.ListenAndServe(addr, context.ClearHandler(auth0Proxy))
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
