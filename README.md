# The de Editor

de is a programmer's editor, where that programmer happens to be [driusan](https://github.com/driusan/).

It's kind of like a bastard child of vim and Plan 9's ACME editor, because vim feels inadequate on a 
computer with a mouse after using acme, and acme feels inadequate on a computer with a keyboard after 
using vi.

Like vim, it's a modal editor with syntax highlighting that uses hjkl for movement.
Like acme, it attempts to exploit your current environment instead of replacing it and tries to
make the mouse useful.

See [USAGE.md](USAGE.md) for usage instructions.

## Features

* Syntax highlighting (currently Go only)
* vi-like keybindings and philosophy
* acme-like mouse bindings and philosophy


## Limitations and Bugs

* vi-like functionality not fully implemented (most notably and some movement verbs like '%' are missing.)
* Can not open multiple files/windows at a time. (if your workflow is like mine, it means you often
  save and quit, do something in the shell, and then relaunch your editor. The startup time should
  be fast enough to support this style of workflow.)
* Missing acme style window tag to use as a scratch space.

# Installation

It should be installable with the standard go tools:

```
go get github.com/driusan/de
```

