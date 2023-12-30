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
