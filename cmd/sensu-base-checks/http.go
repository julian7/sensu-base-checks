package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/hako/durafmt"
	"github.com/julian7/sensulib"
	"github.com/karrick/tparse"
	"github.com/spf13/cobra"
)

type httpConfig struct {
	URL       string
	Timeout   string
	timeout   time.Duration
	Headers   []string
	Insecure  bool
	Certfile  string
	CAfile    string
	Expiry    string
	expiry    time.Time
	Method    string
	Response  uint
	Redirect  string
	UserAgent string
	Data      string
}

func httpCmd() *cobra.Command {
	config := &httpConfig{}
	cmd := sensulib.NewCommand(
		config,
		"http",
		"HTTP check",
		`Checks for HTTP services

This check runs a HTTP query, and inspects return values. Returns

- Unknown on configuration issues,
- Warning on nearing TLS cert expiry or not matching, but non-error HTTP codes,
- Critical on any other cases.

Timeout duration can be provided in short range (eg. ms, s, m, h), cert expiry
can be provided with longer range too (like d, w, mo).
`,
	)
	flags := cmd.Flags()
	flags.StringVarP(&config.URL, "url", "u", "http://127.0.0.1:80/", "Target URL")
	flags.StringVarP(&config.Timeout, "timeout", "t", "5s", "Connection timeout")
	flags.StringSliceVarP(&config.Headers, "header", "H", []string{}, "HTTP header")
	flags.BoolVarP(&config.Insecure, "insecure", "k", false, "Enable insecure connections")
	flags.StringVarP(&config.Certfile, "cert", "c", "", "Certificate file")
	flags.StringVarP(&config.CAfile, "ca", "C", "", "CA Certificate file")
	flags.StringVarP(&config.Expiry, "expiry", "e", "", "Warn EXPIRY before cert expires (duration, like 5d)")
	flags.StringVarP(&config.Method, "method", "X", "GET", "HTTP method")
	flags.StringVarP(&config.UserAgent, "user-agent", "A", "", "User agent")
	flags.StringVarP(&config.Data, "body", "d", "", "HTTP body")
	flags.UintVarP(&config.Response, "response", "r", 2, "HTTP error code to expect; use 3-digits for exact, 1-digit for first digit check")
	flags.StringVarP(&config.Redirect, "redirect", "R", "", "Expect redirection to")

	return cmd
}

func (conf *httpConfig) check() error {
	var err error

	conf.timeout, err = time.ParseDuration(conf.Timeout)
	if err != nil {
		return fmt.Errorf("cannot parse --timeout: %w", err)
	}

	if len(conf.Expiry) != 0 {
		var err error

		//conf.expiry, err = time.ParseDuration(conf.Expiry)
		conf.expiry, err = tparse.ParseNow(time.RFC3339, "now+"+conf.Expiry)
		if err != nil {
			return fmt.Errorf("cannot parse --expiry: %w", err)
		}
	}

	tests := []struct {
		opt   string
		check bool
		err   string
	}{
		{
			"response",
			conf.Response < 1 ||
				(conf.Response > 5 && conf.Response < 100) ||
				conf.Response > 599,
			"should be between 1 and 5, or 100 and 599",
		},
		{
			"response",
			len(conf.Redirect) != 0 &&
				!(conf.Response == 3 || (conf.Response >= 300 && conf.Response < 400)),
			"should expect 3xx if redirect is also expected",
		},
	}
	for _, test := range tests {
		if test.check {
			return fmt.Errorf("--%s %s", test.opt, test.err)
		}
	}

	return nil
}

func (conf *httpConfig) Run(cmd *cobra.Command, args []string) error {
	if err := conf.check(); err != nil {
		return sensulib.Unknown(err)
	}

	tlsconfig := &tls.Config{}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: conf.timeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsconfig,
		},
	}

	req, err := http.NewRequest(conf.Method, conf.URL, strings.NewReader(conf.Data))
	if err != nil {
		return sensulib.Unknown(fmt.Errorf("cannot assemble HTTP request: %w", err))
	}

	if len(conf.Headers) > 0 {
		for _, item := range conf.Headers {
			items := strings.SplitN(item, ":", 2)
			req.Header.Set(items[0], strings.Trim(items[1], " \t\r\n"))
		}
	}

	if len(conf.UserAgent) != 0 {
		req.Header.Set("User-Agent", conf.UserAgent)
	}

	if conf.Insecure {
		tlsconfig.InsecureSkipVerify = true
	}

	var certExpiresAt time.Time

	if len(conf.Expiry) != 0 {
		tlsconfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			certExpiresAt = verifiedChains[0][0].NotAfter
			return nil
		}
	}

	if len(conf.Certfile) != 0 {
		cert, err := tls.LoadX509KeyPair(conf.Certfile, conf.Certfile)
		if err != nil {
			return sensulib.Unknown(fmt.Errorf("cannot load --certfile: %w", err))
		}

		tlsconfig.Certificates = []tls.Certificate{cert}
	}

	if len(conf.CAfile) != 0 {
		cacontents, err := ioutil.ReadFile(conf.CAfile)
		if err != nil {
			return sensulib.Unknown(fmt.Errorf("cannot load --ca: %w", err))
		}

		certpool := x509.NewCertPool()
		certpool.AppendCertsFromPEM(cacontents)
	}

	resp, err := client.Do(req)
	if err != nil {
		return sensulib.Crit(err)
	}

	defer resp.Body.Close()

	// expiry
	if len(conf.Expiry) != 0 {
		if certExpiresAt.Before(conf.expiry) {
			return sensulib.Warn(fmt.Errorf(
				"certificate will expire in %s",
				durafmt.Parse(time.Until(certExpiresAt)).LimitFirstN(4).String(),
			))
		}
	}
	// response
	sfx := ""
	response := uint(resp.StatusCode)

	if conf.Response < 10 {
		response = response / 100
		sfx = "xx"
	}

	if conf.Response != response {
		err := fmt.Errorf(
			"returned with code %s, expected %d%s",
			resp.Status,
			conf.Response,
			sfx,
		)

		if conf.Response >= 400 {
			return sensulib.Crit(err)
		}

		return sensulib.Warn(err)
	}

	// redirect
	redirect := resp.Header.Get("Location")
	if len(conf.Redirect) != 0 && len(redirect) != 0 {
		if redirect != conf.Redirect {
			return sensulib.Crit(fmt.Errorf(
				"redirected to %s, expected to %s",
				redirect,
				conf.Redirect,
			))
		}
	} else if len(conf.Redirect) != 0 {
		return sensulib.Crit(fmt.Errorf("not redirected to %s", conf.Redirect))
	} else if len(redirect) != 0 {
		return sensulib.Crit(fmt.Errorf("unexpected redirection to %s", redirect))
	}

	return nil
}
