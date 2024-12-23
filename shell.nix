
{ pkgs ? import <nixpkgs> {} }:

let
in pkgs.mkShell {

	buildInputs = [
		pkgs.go
	];

	shellHook = ''
		export PATH="$(cd ./bootstrap/ && pwd):$PATH"
	'';

}
