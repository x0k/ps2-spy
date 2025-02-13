{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-24.11";
    nixpkgsUnstable.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    mk.url = "github:x0k/mk";
  };
  outputs =
    {
      self,
      nixpkgs,
      nixpkgsUnstable,
      mk,
    }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
      unstablePkgs = import nixpkgsUnstable { inherit system; };
      gotext = pkgs.buildGoModule rec {
        pname = "gotext";
        version = "0.22.0";
        src = pkgs.fetchFromGitHub {
          owner = "golang";
          repo = "text";
          rev = "v${version}";
          sha256 = "sha256-m8LVnzj+VeclJflfgO7UcOSYSS052RvRgyjTXCgK8As=";
        };
        vendorHash = "sha256-e5DoFMRu3uWQeeWAVd18/nLXOEAfXBRmrH/laWf7C/Y=";
        subPackages = [ "cmd/gotext" ];
      };
    in
    {
      devShells.${system} = {
        default = pkgs.mkShell {
          buildInputs = [
            mk.packages.${system}.default
            unstablePkgs.go_1_24
            pkgs.air
            pkgs.go-migrate
            pkgs.golangci-lint
            pkgs.sqlc
            gotext
            pkgs.gotests
            pkgs.delve
          ];
          shellHook = ''
            source <(COMPLETE=''${SHELL##*/} mk)
          '';
          # CGO_CFLAGS="-U_FORTIFY_SOURCE -Wno-error";
          # CGO_CPPFLAGS="-U_FORTIFY_SOURCE -Wno-error";
        };
      };
    };
}
