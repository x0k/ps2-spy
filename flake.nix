{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-25.11";
    nixpkgs-unstable.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    mk.url = "github:x0k/mk";
  };
  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-unstable,
      mk,
    }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
      unstablePkgs = import nixpkgs-unstable { inherit system; };
      gotext = pkgs.buildGoModule.override { go = pkgs.go_1_26; } rec {
        pname = "gotext";
        version = "0.37.0";
        src = pkgs.fetchFromGitHub {
          owner = "golang";
          repo = "text";
          rev = "v${version}";
          sha256 = "sha256-y2XH1cYuj+xGpRgtkYNIwKnsu/L6XwUYnAFMrJyvqIA=";
        };
        vendorHash = "sha256-ayEih1Px5wTwCm5dJ2oKexPB+NikKMdPy+4wCni1Bg0=";
        subPackages = [ "cmd/gotext" ];
      };
    in
    {
      devShells.${system} = {
        default = pkgs.mkShell {
          buildInputs = [
            mk.packages.${system}.default
            pkgs.go_1_26
            pkgs.air
            pkgs.go-migrate
            pkgs.go-mockery
            unstablePkgs.golangci-lint
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
