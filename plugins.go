package main

import (
	// Default command plugins. This should probably always be here.
	_ "github.com/driusan/de/actions/defaults"
	// More exotic optional plugins.
	_ "github.com/driusan/de/plugins"

	// Renderer to view PNG or JPEG images
	_ "github.com/driusan/de/renderer/imagerenderer"

	// Syntax highlighting plugins
	_ "github.com/driusan/de/renderer/gorenderer"
	_ "github.com/driusan/de/renderer/markdown"
	_ "github.com/driusan/de/renderer/phprenderer"
)
