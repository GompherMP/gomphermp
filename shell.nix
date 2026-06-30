{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
	buildInputs = [ 
    pkgs.go
    pkgs.gopls
    pkgs.typst
    pkgs.tinymist
    pkgs.python312
    pkgs.python312Packages.matplotlib
  ];
}
