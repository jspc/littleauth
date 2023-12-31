[![Go Report Card](https://goreportcard.com/badge/github.com/jspc/littleauth)](https://goreportcard.com/report/github.com/jspc/littleauth)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=jspc_littleauth&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=jspc_littleauth)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=jspc_littleauth&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=jspc_littleauth)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=jspc_littleauth&metric=bugs)](https://sonarcloud.io/summary/new_code?id=jspc_littleauth)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=jspc_littleauth&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=jspc_littleauth)
[![Coverage Status](https://coveralls.io/repos/github/jspc/littleauth/badge.svg?branch=main)](https://coveralls.io/github/jspc/littleauth?branch=main)

# littleauth

Teeny tiny little auth gateway dealy.

It:

1. Stores passwords in a `passwd` style file, which can be updated on the fly
1. Reads login forms from a directory somewhere, which require a restart to update
1. Supports virtual hosts and multiple login portals
1. Exposes a path at `/api/v1/auth` which returns either a 302 (to the login form) or a 200
1. Exposes a path at `/api/v1/login` which your login form POSTs to; it returns a 403 (if creds wrong) or returns a cookie (which `/api/v1/auth` routes on) and a 302 to the service requested
1. Exposes a path at `/api/v1/logout` which removes a user's session and returns them back to login

## Goals

* Small, tiny, static, stripped binary
* Copious tests
* No/ few allocations
* Fuck all logic

## Configuration

This project uses a toml file as per:

```toml
[accounts.example.net]
templates = "/www/www.example.net/login"
passwd    = "/www/www.example.net/.htpasswd"
redirect  = "https://accounts.example.new/"
domain    = "example.net"
origins  = ["mail.example.com", "www.example.com", "example.com"]
```

The format of this file is:

```golang

type Config map[string]VirtualHost

type VirtualHost struct {
    TemplateDir  string `toml:"templates"`
    PasswdFile   string `toml:"passwd"`
    Redirect     string `toml:"redirect"`
    TOTPFile     string `toml:"totp"`
    CookieDomain string `toml:"domain"`

    // Origins contains the permitted x-forwarded-host values
    // allowed to authenticate against this virtual host
    Origins []string `toml:"origins"`
}
```

`VirtualHosts` are matched based on the value of the `X-Forwarded-Host` header. Where this header doesn't exist, or there's no match, nothing happens.
