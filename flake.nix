{
  description = "A basic gomod2nix flake";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.gomod2nix.url = "github:nix-community/gomod2nix";

  outputs = { self, nixpkgs, flake-utils, gomod2nix }:
    (flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ gomod2nix.overlays.default ];
          };

          app = pkgs.callPackage ./. { };

          dockerImage = pkgs.dockerTools.buildLayeredImage {
            name = "pbin";
            tag = "latest";
            config.Cmd = [ "${app}/bin/pbin" ];
            config.Env = [ "SSL_CERT_FILE=/etc/ssl/certs/ca-bundle.crt" ];
            contents = [ pkgs.cacert ];
          };

        in
        {
          packages.default = app;
          packages.dockerImage = dockerImage;
          devShells.default = import ./shell.nix { inherit pkgs; };
        })
    );
}
