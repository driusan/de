package main

import (
	// Default command plugins
	_ "github.com/driusan/de/actions/defaults"

	// Syntax highlighting plugins
	_ "github.com/driusan/de/renderer/gorenderer"
	_ "github.com/driusan/de/renderer/markdown"
	_ "github.com/driusan/de/renderer/phprenderer"
)
