{
  description = "KDB Befordringsfradrag Service Unified Workspace Environment";

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
            go
            gopls
            golangci-lint
            sqlite
            nodejs
            typescript-language-server
            prettier
          ];

          shellHook = ''
            export GOPATH="$PWD/.go"
            export PATH="$GOPATH/bin:$PATH"
            echo "🚀 KDB Full Workspace Dev Shell ready"
            echo "   Go:   $(go version)"
            echo "   Node: $(node -v)"
          '';
        };
      }
    );
}
