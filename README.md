# littleauth

Teeny tiny little auth gateway dealy.

It:

1. Stores passwords in a `passwd` style file, which can be updated on the fly
2. Reads login forms from a directory somewhere, which require a restart to update
3. Exposes a path at `/api/v1/auth` which returns either a 302 (to the login form) or a 200
4. Exposes a path at `/api/v1/login` which your login form POSTs to; it returns a 403 (if creds wrong) or returns a cookie (which `/api/v1/auth` routes on) and a 302 to the service requested

## Goals

* Small, tiny, static, stripped binary
* Copious tests
* No/ few allocations
* Fuck all logic

## Configuration

This project uses a toml file as per:

```toml
['*']
templates = "/www/default-form"
passwd    = "/www/.htpasswd"
redirect  = "https://example.com/dashboard"
totp      = "/www/.totp"

[accounts.example.net]
templates = "/www/www.example.net/login"
passwd    = "/www/www.example.net/.htpasswd"
redirect  = "https://accounts.example.new/"
domain    = "accounts.example.com"
```

The format of this file is:

```golang

type Config map[string]VirtualHost

type VirtualHost struct {
    TemplateDir  string `toml:"templates"`
    PasswdFile   string `toml:"passwd"`
    Redirect     string `toml:"redirect"`
    TOTPFile     string `toml:"totp,omitempty"`
    CookieDomain string `toml:"domain,omitempty"`
}
```

`VirtualHosts` are matched based on the value of the `X-Forwarded-Host` header. Where this header doesn't exist, or there's no match, the default virtualhost matches.

If the config file doesn't contain a default virtualhost then the app doesn't start.
