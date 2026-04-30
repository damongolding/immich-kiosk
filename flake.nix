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
        version = "0.38.0"; # hard coded for now

        nodeModulesHashes = {
          "x86_64-linux"  = pkgs.lib.fakeHash; # generate on an x86_64-linux machine
          "aarch64-linux" = pkgs.lib.fakeHash; # generate on an aarch64-linux machine
          "x86_64-darwin" = pkgs.lib.fakeHash; # generate on an x86_64 Mac
          "aarch64-darwin" = "sha256-67fg+bPFMfUjCNbxUMkc8r2HDBK2g6aE6CaWoOnX/B4="; # generate on an Apple Silicon Mac
        };

        # Step 1: install frontend deps
        node_modules = pkgs.stdenv.mkDerivation {
          pname = "immich-kiosk-node_modules";
          inherit version;
          nativeBuildInputs = [ pkgs.bun ];
          phases = [ "buildPhase" "installPhase" ];

          impureEnvVars = pkgs.lib.fetchers.proxyImpureEnvVars ++
            [ "GIT_PROXY_COMMAND" "SOCKS_SERVER" ];

          buildPhase = ''
            cp ${./frontend/package.json} package.json
            cp ${./frontend/bun.lock} bun.lock
            bun install --no-progress --frozen-lockfile
          '';

          installPhase = ''
            mkdir -p $out
            cp -R ./node_modules $out
          '';

          outputHash     = nodeModulesHashes.${system} or (throw "Unsupported system: ${system}");
          outputHashAlgo = "sha256";
          outputHashMode = "recursive";
        };

        # Step 2: build frontend assets
        frontend = pkgs.stdenv.mkDerivation {
          pname = "immich-kiosk-frontend";
          inherit version;
          src = ./frontend;
          nativeBuildInputs = [ pkgs.bun pkgs.nodejs ];

          configurePhase = ''
            cp -R ${node_modules}/node_modules .
          '';

          buildPhase = ''
            bun run css
            bun run js
            bun run url-builder
          '';

          installPhase = ''
            mkdir -p $out
            cp -r ./public/assets $out
          '';
        };

      in
      {
        packages = {
          default = self.packages.${system}.immich-kiosk;

          immich-kiosk = pkgs.buildGoModule {
            pname = "immich-kiosk";
            inherit version;
            src = ./.;
            vendorHash = "sha256-O1cH0EGHdOgpo+zhdlYFKVK4cOHg8ZKpsJezdKBv+K0=";

            nativeBuildInputs = with pkgs; [ go-task ];

            preBuild = ''

              # Satisfy go:embed frontend/public
              mkdir -p frontend/public/assets
              cp -r ${frontend}/assets/. frontend/public/assets/


              # Generate templ templates
              go tool templ generate
            '';

            ldflags = [
              "-s" "-w"
              "-X main.version=${version}"
            ];

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
            golangci-lint
            bun
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
