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
      rec {
        packages = rec {
          # 1. Frontend UI build (Vite + React)
          # Update hash with: nix run nixpkgs#prefetch-npm-deps ui/package-lock.json
          ui = pkgs.buildNpmPackage {
            pname = "kdb-ui";
            version = "1.0.0";
            src = ./ui;
            npmDepsHash = "sha256-qVDUQL8u6j1kLnG71mdFkcayjxcqTCBD3UnnNVJ4PMA=";
            installPhase = ''
              mkdir -p $out
              cp -r dist/* $out/
            '';
          };

          # 2. Native Go Service with embedded UI assets
          # Update vendorHash with: set to pkgs.lib.fakeHash and read mismatch
          kdb = pkgs.buildGoModule {
            pname = "kdb-service";
            version = "1.0.0";
            src = ./service;
            vendorHash = "sha256-WpGhvUD7Oq8+kx+HjG2ocVYyzZVPMIFwBoILjcp7X8o=";
            env.CGO_ENABLED = 0;
            subPackages = [ "cmd/server" ];
            ldflags = [ "-s" "-w" ];
            preBuild = ''
              mkdir -p cmd/server/dist
              rm -rf cmd/server/dist/*
              cp -r ${ui}/* cmd/server/dist/
            '';
            postInstall = ''
              mv $out/bin/server $out/bin/kdb-service
            '';
          };

          # 3. ARM64 cross-compiled binary for Oracle Cloud Infrastructure (OCI)
          kdb-arm64 = pkgs.buildGoModule {
            pname = "kdb-service-arm64";
            version = "1.0.0";
            src = ./service;
            vendorHash = "sha256-WpGhvUD7Oq8+kx+HjG2ocVYyzZVPMIFwBoILjcp7X8o=";
            preBuild = ''
              mkdir -p cmd/server/dist
              rm -rf cmd/server/dist/*
              cp -r ${ui}/* cmd/server/dist/
            '';
            buildPhase = ''
              runHook preBuild
              mkdir -p $out/bin
              CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $out/bin/kdb-service-linux-arm64 ./cmd/server
              runHook postBuild
            '';
            doCheck = false;
          };

          default = kdb;
        };

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
