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
        packages = {
          default = self.packages.${system}.immich-kiosk;

          immich-kiosk = pkgs.buildGoModule rec {
            pname = "immich-kiosk";

            src = ./.;

            vendorHash = pkgs.lib.hashFile "sha256-";

            nativeBuildInputs = with pkgs; [
              nodePackages.pnpm
              go-task
            ];

            # Ensure embedded assets are included
            subPackages = [ "." ];

            meta = with pkgs.lib; {
              description = "Highly configurable slideshows for displaying Immich assets on browsers and devices.";
              homepage = "https://github.com/damongolding/immich-kiosk";
              license = licenses.agpl3Only;
              mainProgram = "immich-kiosk";
            };
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-task
            nodePackages.pnpm
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
