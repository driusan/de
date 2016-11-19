package main

import (
	// Default command plugins. This should probably always be here.
	_ "github.com/driusan/de/actions/defaults"
	// More exotic optional plugins.
	_ "github.com/driusan/de/plugins/shell"
	//_ "github.com/driusan/de/plugins/redmine"
	// RENDERERS:
	// The order here matters. The most recently registered (lowest down the list)
	// renderer will win.

	// Hex mode renderer should come before the default renderer, since
	// it claims it can render anything. We want the default to be plain
	// text, not a hex dump.
	_ "github.com/driusan/de/renderer/hex"
	// The default plain text renderer
	_ "github.com/driusan/de/renderer/nosyntax"

	// Renderer to view PNG or JPEG images
	_ "github.com/driusan/de/renderer/imagerenderer"

	// Syntax highlighting plugins.
	_ "github.com/driusan/de/renderer/gorenderer"
	_ "github.com/driusan/de/renderer/html"
	_ "github.com/driusan/de/renderer/jsrenderer"
	_ "github.com/driusan/de/renderer/markdown"
	_ "github.com/driusan/de/renderer/phprenderer"
)
