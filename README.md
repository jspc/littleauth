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
