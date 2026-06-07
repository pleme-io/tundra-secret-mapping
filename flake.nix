# flake.nix — tundra-secret-mapping (GSDS Biblioteca) via substrate's go-library-flake.
# vendorHash OMITTED → spec-sourced (__from-spec__); clean nix build lands once
# errors-go is published. Pre-publish proof is `go test` (green).
{
  description = "tundra-secret-mapping — typed JSON-secret field-mapping with per-secret overrides";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    substrate = {
      url = "github:pleme-io/substrate";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs @ { self, nixpkgs, substrate, ... }:
    (import substrate.goLibraryFlakeBuilder { inherit nixpkgs; }) {
      name = "tundra-secret-mapping";
      version = "0.1.0";
      src = self;
      repo = "pleme-io/tundra-secret-mapping";
    };
}
