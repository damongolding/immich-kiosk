{
  description = "immich-kiosk - Highly configurable slideshows for displaying Immich assets on browsers and devices.";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-task
            golangci-lint
            nodejs
          ];
          shellHook = ''
            echo "immich-kiosk development environment"
            echo "Common commands:"
            echo "  task --list    - Show available tasks"
            echo "  task dev       - Start development server"
            echo "  task build     - Build the application"
          '';
        };
      }
    );
}
