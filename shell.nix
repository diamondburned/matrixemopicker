{ unstable ? import <unstable> {} }:

unstable.stdenv.mkDerivation rec {
	name = "cchat-gtk";
	version = "0.0.2";

	buildInputs = with unstable; [
		gnome3.glib
		gnome3.gtk
	];

	nativeBuildInputs = with unstable; [
		pkgconfig
		go
		wrapGAppsHook
	];
}
