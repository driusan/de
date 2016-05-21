# The de Editor

de is a programmer's editor. (Where that programmer happens to be [driusan](https://github.com/driusan/).)

It's kind of like a bastard child of vim and Plan 9's acme editor, because vim feels inadequate on a 
computer with a mouse after using acme, and acme feels inadequate on a computer with a keyboard after 
using vi.

Like vim, it's a modal editor with syntax highlighting that uses hjkl for movement.
Like acme, it attempts to exploit your current OS environment instead of replacing it and tries to
make the mouse useful.

See [USAGE.md](USAGE.md) for usage instructions.

![de screenshot](https://driusan.github.io/de/descreenshot_readme.png)
## Features

* Syntax highlighting (currently only Go and PHP with some basic markdown.)
* vi-like keybindings and philosophy.
* acme-like mouse bindings and philosophy.
* Ability to write plugins in Go. See [PLUGINS.md](PLUGINS.md).

![de screenshot](https://driusan.github.io/de/descreenshot_code.png)
## Limitations and Bugs

* vi-like functionality not fully implemented (most notably some movement verbs like '%' are missing,
  and see notes in issue #1.)
* Can not open multiple files/windows at a time. (if your workflow is like mine, it means you often
  save and quit, do something in the shell, and then relaunch your editor. The startup time should
  be fast enough to support this style of workflow.)

# Installation

It should be installable with the standard go tools:

```
go get github.com/driusan/de
```

Then as long as $GOPATH/bin is in your path, you can launch with `de [filename]`

