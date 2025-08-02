{
  description = "A basic gomod2nix flake";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.gomod2nix.url = "github:nix-community/gomod2nix";

  outputs = { self, nixpkgs, flake-utils, gomod2nix }:
    (flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };

        pnpm = pkgs.pnpm_10;

        # Build the React frontend
        frontend = pkgs.stdenv.mkDerivation (finalAttrs: {
          pname = "pbin-frontend";
          version = "1.0.0";

          src = ./.;

          nativeBuildInputs = [ pkgs.nodejs pnpm pnpm.configHook ];

          pnpmDeps = pnpm.fetchDeps {
            inherit (finalAttrs) pname version src;
            fetcherVersion = 2;
            hash = "sha256-LlThF3T6MMOT1gxFsfXULFUi3GnaVnFTNEWq+3Go0Hw=";
          };

          buildPhase = ''
            pnpm generate-api
            pnpm build
          '';

          installPhase = ''
            mkdir -p $out
            cp -r static/* $out/
          '';

        });

        # Build the Go backend with embedded frontend
        app = pkgs.buildGoApplication {
          pname = "pbin";
          version = "0.1";
          pwd = ./.;
          src = ./.;
          modules = ./gomod2nix.toml;

          # Embed the frontend build into the Go binary
          preBuild = ''
            mkdir -p static
            cp -r ${frontend}/* static/
          '';
        };

        dockerImage = pkgs.dockerTools.buildLayeredImage {
          name = "pbin";
          tag = "latest";
          config.Cmd = [ "${app}/bin/pbin" ];
          config.Env = [ "SSL_CERT_FILE=/etc/ssl/certs/ca-bundle.crt" ];
          contents = [ pkgs.cacert ];
        };

      in {
        packages = {
          default = app;
          frontend = frontend;
          dockerImage = dockerImage;
        };
        devShells.default = import ./shell.nix { inherit pkgs; };
      }));
}
