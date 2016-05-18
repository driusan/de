package main

import (
	// Default command plugins
	_ "github.com/driusan/de/actions/defaults"

	// Go syntax highlighting plugin
	_ "github.com/driusan/de/renderer/gorenderer"
	//_ "github.com/driusan/de/renderer/markdown"
)
