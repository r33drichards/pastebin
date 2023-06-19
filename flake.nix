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
            extraCommands = ''
              mkdir -p /etc/ssl/certs
              cp ${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt /etc/ssl/certs/
            '';
          };
        in
        {
          packages.default = app;
          packages.dockerImage = dockerImage;
          devShells.default = import ./shell.nix { inherit pkgs; };
        })
    );
}
