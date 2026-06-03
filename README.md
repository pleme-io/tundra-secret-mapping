# tundra-secret-mapping

A typed JSON-secret field-mapping with per-secret overrides — project a JSON
secret onto canonical credential fields, once, for the whole fleet.

## What

A `Mapping` declares which JSON keys of a secret supply the canonical credential
`Field`s (`username`, `password`, `private_key`, `passphrase`, …). A `Default`
field map applies to every secret; per-secret-path `Overrides` shadow the
default per-field. `Apply(path, secretJSON)` returns the projected
`map[Field]string`. It is a pure, generic field-mapping primitive: it models any
JSON secret and names no particular project — the concrete mapping values a
deployment uses are its own data.

## Why

Integration adapters (ServiceNow, ESO, UI rotators) each otherwise fork the same
"pull username/password/key out of this secret JSON" logic in a different
language. tundra-secret-mapping is the one shared, typed mapping so every adapter
projects secrets the same way, with per-secret overrides for the inevitable
odd-shaped secret.

## Install

```
go get github.com/pleme-io/tundra-secret-mapping
```

Nix (via substrate):

```nix
outputs = { self, nixpkgs, substrate, ... }:
  (import substrate.goLibraryFlakeBuilder { inherit nixpkgs; }) {
    name = "tundra-secret-mapping"; version = "0.1.0"; src = self;
  };
```

## Usage

Built on: [errors-go] (typed, code-carrying errors).

```go
m := secretmapping.New(
    secretmapping.WithDefault(secretmapping.FieldMap{
        secretmapping.Username: "user",
        secretmapping.Password: "pass",
    }),
    secretmapping.WithOverride("/ssh/host", secretmapping.FieldMap{
        secretmapping.Username:   "login",
        secretmapping.PrivateKey: "key",
    }),
)

creds, err := m.Apply("/ssh/host", secretJSON)
if err != nil { return errs.Exit(err) }
// creds[secretmapping.Username], creds[secretmapping.PrivateKey], …
```

## Configuration

The `Mapping` is itself the configuration shape — a consumer loads its
`Default`/`Overrides` from a configured source via `shikumi-go` and constructs
the `Mapping` from the loaded typed values. The library performs no config
loading of its own.

## Release

Pull-model (Go modules): an annotated `vX.Y.Z` tag is the release;
`proxy.golang.org` + pkg.go.dev index it. See the GSDS module delivery FSM.

[errors-go]: https://github.com/pleme-io/errors-go
