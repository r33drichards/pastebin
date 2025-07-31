{ pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
    in
    import (fetchTree nixpkgs.locked) {
      overlays = [
        (import "${fetchTree gomod2nix.locked}/overlay.nix")
      ];
    }
  )
}:

let
  goEnv = pkgs.mkGoEnv { pwd = ./.; };
in
pkgs.mkShell {
  hardeningDisable = [ "fortify" ];

  packages = [
    goEnv
    pkgs.gomod2nix
    pkgs.just
    pkgs.terraform
    pkgs.awscli
    pkgs.git
    pkgs.nixfmt
    pkgs.golangci-lint
    pkgs.go-tools
    pkgs.gotools
    pkgs.gopls
    pkgs.go-outline
    pkgs.gopkgs
    pkgs.gocode-gomod
    pkgs.godef
    pkgs.golint
    pkgs.delve
    pkgs.nixpkgs-fmt
    pkgs.nodejs_20
    pkgs.nodePackages.pnpm
  ];
}
