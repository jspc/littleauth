package main

import (
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/memory"
	"github.com/tg123/go-htpasswd"
)

var MissingDefaultVHostErr = errors.New("missing default virtual host")

// Config maps a set of virtualhosts with host names
type Config map[string]*VirtualHost

// VirtualHost contains the various configurables for specific
// virtual hosts, such as various auth stuff, or whatever
type VirtualHost struct {
	TemplateDir  string `toml:"templates"`
	PasswdFile   string `toml:"passwd"`
	Redirect     string `toml:"redirect"`
	TOTPFile     string `toml:"totp,omitempty"`
	CookieDomain string `toml:"domain,omitempty"`

	passwd    *htpasswd.File     `toml:"-"`
	templates *template.Template `toml:"-"`
	sm        *session.Session   `toml:"-"`
}

// ReadConfig takes a config file, and organises the various
// virtual hosts and what have you.
//
// This function errors when:
//
//  1. The server config file does not exist
//  2. VHost configurations are wrong, such as missing htpasswd or templates
//  3. The config file doesn't contain a default vhost
func ReadConfig(fn string) (c Config, err error) {
	_, err = toml.DecodeFile(fn, &c)
	if err != nil {
		return
	}

	for _, vh := range c {
		err = vh.Configure()
		if err != nil {
			return
		}
	}

	if _, ok := c["*"]; !ok {
		err = MissingDefaultVHostErr
	}

	return
}

// MatchVHost returns either a names vhost or the default vhost, depending
// on whether the vhost exists
func (c Config) MatchVHost(host []byte) (vh *VirtualHost) {
	vh, ok := c[string(host)]
	if !ok {
		vh = c["*"]
	}

	return
}

// Configure will configure the specfied vhost, with an htpasswd matcher,
// a set of templates, and a session manager
func (vh *VirtualHost) Configure() (err error) {
	vh.passwd, err = htpasswd.New(vh.PasswdFile, htpasswd.DefaultSystems, nil)
	if err != nil {
		return
	}

	vh.templates, err = template.New("login").ParseGlob(filepath.Join(vh.TemplateDir, "*.html.tmpl"))
	if err != nil {
		return
	}

	cfg := session.NewDefaultConfig()
	cfg.CookieName = fmt.Sprintf("%s_%s", vh.CookieDomain, vh.Redirect)
	cfg.Domain = vh.CookieDomain
	cfg.Expiration = time.Second * 604800
	cfg.Secure = true

	cfg.EncodeFunc = session.MSGPEncode
	cfg.DecodeFunc = session.MSGPDecode

	vh.sm = session.New(cfg)

	provider, err := memory.New(memory.Config{})
	if err != nil {
		return
	}

	return vh.sm.SetProvider(provider)
}
