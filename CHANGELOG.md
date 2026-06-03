# Changelog

All notable changes to this project are documented here.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-06-03

### Added
- Initial typed `Mapping` — projects a JSON secret onto canonical credential
  `Field`s (username/password/private_key/passphrase) via a `Default` field map
  plus per-secret-path `Overrides` (override entries shadow defaults per-field).
  Options-built `New`, `Validate`, and `Apply` with code-carrying errors via
  `errors-go`. Names no particular project.
