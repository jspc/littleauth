package main

import (
	"context"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"golang.org/x/net/publicsuffix"
)

var testConfig = &Config{
	"auth.example.com": &VirtualHost{
		TemplateDir:  "testdata/form",
		PasswdFile:   "testdata/passwd",
		Redirect:     "http://example.com",
		CookieDomain: "example.com",
		Origins:      []string{"example.com", "mail.example.com"},
	},

	"dodgy-sm-get.example.com": &VirtualHost{
		TemplateDir:  "testdata/form",
		PasswdFile:   "testdata/passwd",
		Redirect:     "http://example.com",
		CookieDomain: "example.com",
		Origins:      []string{"d1.example.com"},
	},

	"dodgy-sm-save.example.com": &VirtualHost{
		TemplateDir:  "testdata/form",
		PasswdFile:   "testdata/passwd",
		Redirect:     "http://example.com",
		CookieDomain: "example.com",
		Origins:      []string{"d2.example.com"},
	},
}

func init() {
	for _, vh := range *testConfig {
		err := vh.Configure()
		if err != nil {
			panic(err)
		}
	}

	(*testConfig)["dodgy-sm-get.example.com"].sm = &dummySessionManager{getErr: true}
	(*testConfig)["dodgy-sm-save.example.com"].sm = &dummySessionManager{saveErr: true}
}

func newTestServerClient() *http.Client {
	s, _ := New(testConfig)
	srv := &fasthttp.Server{
		Handler:         s.Handler,
		ReadBufferSize:  8 * 1024,
		WriteBufferSize: 8 * 1024,
	}

	ln := fasthttputil.NewInmemoryListener()
	go srv.Serve(ln)

	j, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		panic(err)
	}

	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return ln.Dial()
			},
		},
		Timeout: time.Second,
		Jar:     j,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func TestNew(t *testing.T) {
	_, err := New(testConfig)
	if err != nil {
		t.Errorf("unexpected error %#v", err)
	}
}

func TestServer_Auth_VHost_Matching(t *testing.T) {
	c := newTestServerClient()

	for _, test := range []struct {
		forwardedHost string
		expectStatus  int
		expectRedir   string
	}{
		{"example.com", 303, "https://auth.example.com"},
		{"mail.example.com", 303, "https://auth.example.com"},
		{"www.example.com", 404, ""},
	} {
		t.Run(test.forwardedHost, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "http://auth.example.com/api/v1/auth", nil)
			req.Header.Add("X-Forwarded-Host", test.forwardedHost)

			resp, err := c.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if test.expectStatus != resp.StatusCode {
				t.Errorf("expected status %d, received %d", test.expectStatus, resp.StatusCode)
			}
		})
	}
}

func TestServer_Auth_Missing_Forwarded_Host(t *testing.T) {
	c := newTestServerClient()
	req, _ := http.NewRequest(http.MethodGet, "http://auth.example.com/api/v1/auth", nil)

	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if 404 != resp.StatusCode {
		t.Errorf("expected status %d, received %d", 404, resp.StatusCode)
	}
}

func TestServer_Auth_With_Session_Manager_Errors(t *testing.T) {
	c := newTestServerClient()

	req, _ := http.NewRequest(http.MethodGet, "http://auth.example.com/api/v1/auth", nil)
	req.Header.Add("X-Forwarded-Host", "d1.example.com")

	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if 500 != resp.StatusCode {
		t.Errorf("expected status %d, received %d", 500, resp.StatusCode)
	}
}

func TestServer_Auth_Is_Authed(t *testing.T) {
	c := newTestServerClient()

	form := url.Values{}
	form.Add("username", "tests")
	form.Add("password", "teststests")

	req, err := http.NewRequest(http.MethodPost, "http://auth.example.com/api/v1/login", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("X-Forwarded-Host", "auth.example.com")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if 303 != resp.StatusCode {
		t.Errorf("login: expected status %d, received %d", 303, resp.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodGet, "http://auth.example.com/api/v1/auth", nil)
	req.Header.Add("X-Forwarded-Host", "example.com")

	resp, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if 200 != resp.StatusCode {
		t.Errorf("auth: expected status %d, received %d", 200, resp.StatusCode)
	}
}

func TestServer_Login(t *testing.T) {
	c := newTestServerClient()

	for _, test := range []struct {
		name          string
		password      string
		forwardedHost string
		expectStatus  int
	}{
		{"Happy path", "teststests", "auth.example.com", 303},
		{"Empty x-forwarded-host", "teststests", "", 404},
		{"Incorrect x-forwarded-host", "teststests", "example.com", 404},
		{"Incorrect password", "wronggggg", "auth.example.com", 403},
	} {
		t.Run(test.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("username", "tests")
			form.Add("password", test.password)

			req, err := http.NewRequest(http.MethodPost, "http://auth.example.com/api/v1/login", strings.NewReader(form.Encode()))
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Add("X-Forwarded-Host", test.forwardedHost)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			resp, err := c.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if test.expectStatus != resp.StatusCode {
				t.Errorf("login: expected status %d, received %d", test.expectStatus, resp.StatusCode)
			}
		})
	}
}

func TestServer_Login_Missing_Username(t *testing.T) {
	c := newTestServerClient()

	form := url.Values{}
	form.Add("password", "teststests")

	req, err := http.NewRequest(http.MethodPost, "http://auth.example.com/api/v1/login", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("X-Forwarded-Host", "auth.example.com")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if 403 != resp.StatusCode {
		t.Errorf("login: expected status %d, received %d", 303, resp.StatusCode)
	}
}

func TestServer_Login_With_Session_Manager_Errors(t *testing.T) {
	c := newTestServerClient()

	for _, test := range []string{
		"dodgy-sm-get.example.com",
		"dodgy-sm-save.example.com",
	} {
		t.Run(test, func(t *testing.T) {
			form := url.Values{}
			form.Add("username", "tests")
			form.Add("password", "teststests")

			req, err := http.NewRequest(http.MethodPost, "http://auth.example.com/api/v1/login", strings.NewReader(form.Encode()))
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Add("X-Forwarded-Host", test)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			resp, err := c.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if 500 != resp.StatusCode {
				t.Errorf("expected status %d, received %d", 500, resp.StatusCode)
			}
		})
	}
}

func TestServer_Logout(t *testing.T) {
	c := newTestServerClient()

	for _, test := range []struct {
		name          string
		forwardedHost string
		expectStatus  int
	}{
		{"Happy path", "auth.example.com", 303},
		{"Empty x-forwarded-host", "", 404},
		{"Incorrect x-forwarded-host", "example.com", 404},
	} {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://auth.example.com/api/v1/logout", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Add("X-Forwarded-Host", test.forwardedHost)

			resp, err := c.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if test.expectStatus != resp.StatusCode {
				t.Errorf("login: expected status %d, received %d", test.expectStatus, resp.StatusCode)
			}
		})
	}
}

func TestServer_Logout_With_Session_Manager_Errors(t *testing.T) {
	c := newTestServerClient()

	for _, test := range []string{
		"dodgy-sm-get.example.com",
		"dodgy-sm-save.example.com",
	} {
		t.Run(test, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://auth.example.com/api/v1/logout", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Add("X-Forwarded-Host", test)

			resp, err := c.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if 500 != resp.StatusCode {
				t.Errorf("expected status %d, received %d", 500, resp.StatusCode)
			}
		})
	}
}

func TestServer_RenderForm(t *testing.T) {
	c := newTestServerClient()

	for _, test := range []struct {
		name          string
		forwardedHost string
		expectStatus  int
	}{
		{"Happy path", "auth.example.com", 200},
		{"Empty x-forwarded-host", "", 404},
		{"Incorrect x-forwarded-host", "example.com", 404},
	} {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://auth.example.com", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Add("X-Forwarded-Host", test.forwardedHost)

			resp, err := c.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if test.expectStatus != resp.StatusCode {
				t.Errorf("login: expected status %d, received %d", test.expectStatus, resp.StatusCode)
			}
		})
	}
}
