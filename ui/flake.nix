{
  description = "KDB Frontend UI Development Environment";

  inputs = {
    nixpkgs.url = "nixpkgs";
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
            nodejs
            typescript-language-server
            prettier
          ];

          shellHook = ''
            echo "✨ KDB UI Dev Shell ready (Node $(node -v), NPM $(npm -v))"
          '';
        };
      }
    );
}
