# de Plugins

There are two types of plugins that can you might want to add to de: things that add new
commands, and things that change the rendering (ie to add syntax highlighting.)

## Command Plugins

The easiest way to write a "plugin" for de is to just to write a shell command that can be
executed with either a middle click or the enter key to insert into the current buffer,
but you may want to do something more complicated that needs to know the context of the
current buffer.

To add a new "builtin" command plugin, all you need to do is write a Go package
that calls github.com/driusan/de/actions.RegisterAction in your package init function,
add an import to plugins.go, and recompile (but it's Go, so that should be fast and easy.)

See actions/defaults/defaultactions.go for an example, which contains the default
builtins.

actions.RegisterAction takes two arguments: the command name that should be used
to invoke it, and a callback function to execute when the command is invoked.

The callback function should take two parameters:

1. The arguments that the user invoked the command with (not yet implemented)
2. A *demodel.CharBuffer which contains the current character buffer, snarf buffer, and dot.
   See the godoc of the demodel package for more details. You can manipulate the CharBuffer.Buffer
   slice however you want, but the positions package has helpers to calculate positions relative
   to the cursor/selections, and the actions package has helpers to perform common manipulations.

## Rendering Plugins

A rendering plugin changes the way that a character buffer is rendered to the screen. You can
create a renderer by implementing the github.com/driusan/de/renderer/Renderer interface and
registering it with renderer.RegisterRenderer in your init function, then adding it to plugins.go
and recompiling similarly to command plugins.

The Renderer interface requires two functions: a CanRender(demodel.CharBuffer) method which
determines whether de should use your renderer to attempt to render a character buffer or skip
to the next available renderer, and a Render(demodel.CharBuffer) method which returns an image.Image
to be rendered to the screen representing that character buffer, and a renderer.ImageMap which can
be used to convert points in that image into indexes into CharBuffer.Buffer for the mouse to use.

See renderer/gorenderer/gosyntax.go for a sample renderer that supports syntax highlighting for
the Go language.