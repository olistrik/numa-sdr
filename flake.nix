{
  description = "flake shell";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { nixpkgs, ... }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in
    {
      devShells.x86_64-linux = {
        default = pkgs.mkShell {
          packages = with pkgs; [
            git

            go
            # soapysdr
          ];
          # env = {
          #   SOAPY_SDR_PLUGIN_PATH = "${
          #   pkgs.lib.makeSearchPath pkgs.soapysdr.passthru.searchPath [ pkgs.soapyrtlsdr ]
          # }";
          # };
        };
      };
    };
}

