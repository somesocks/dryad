
{ pkgs ? import <nixpkgs> {} }:

let
in pkgs.mkShell {

	buildInputs = [
		pkgs.go
		pkgs.gnumake
	];

	# shellHook = ''
	# '';

}
