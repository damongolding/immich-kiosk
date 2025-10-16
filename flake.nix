{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    {
      self,
      nixpkgs,
    }:
    let
      system = "x86_64-linux";
    in
    {
      devShells.${system}.default =
        let
          pkgs = import nixpkgs {
            inherit system;
          };
        in
        pkgs.mkShellNoCC {
          packages = with pkgs; [
            # Go toolchain
            go
            gotools
            gopls
            go-task

            # Frontend
            nodejs
            pnpm_9

            # Templ for template generation
            templ
          ];
        };
    };
}
