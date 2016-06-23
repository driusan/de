# de Commands

This is a list of commands that can be executed in a default `de` install, 
either from the main buffer or the tag line. See USAGE.md for details about 
executing words.

## File Commands

* `Save`: Saves the current file to the same name it was opened under. (Alias:
	`Put` to make life easier for `acme` users and `w` so that :w works 
	for vi users.)
* `Discard`: Retrieve the current file from the filesystem and discard changes
 	to the buffer. (Alias: `Get` for `acme` users.)
* `Exit`: Quit the current `de` session. Print a warning/error to the tagline
 	instead if the buffer is dirty. (Alias: `Quit` for sanity, and `q` so
 	that :q works for vi users.)
* `ForceQuit`: Discard any changes to the buffer and quit the current session,
 	regardless of if it's dirty or not (Alias: `ForceExit` for symmetry and
 	`q!` for vi users.)
* `SaveExit`: Save any changes and Exit. (Alias: `SaveQuit` for symmetry, and 
	`wq`, `wq!`, and `x` so that an approximation of the commands works for
	vi users.)
* `SaveOrExit`: Save the file if the buffer is dirty, and Quit if the buffer is
	clean. This is mostly to have a command that works the same as the
	escape key. (Alias: `SaveOrQuit` for symmetry.)

## Editing Commands

* `Undo`: Undo the most recent change to the buffer.
* `Paste`: Paste the most recently deleted text from the buffer into the location
	of the cursor, overwriting any selected text (if run from the tagline.)
* `Join`: Joins a number of lines, similar to the 'J' command in vi. If text is
	selected, the selected lines will be joined, otherwise the current
	line will be joined with the next line. If an argument is provided such
	as Join:3 it will join that number of lines (similar to a prefix
	modifier for `J` in vi.)

## Viewport Commands
* `ResetTagline`: Reset the tagline to what it would be if you had just opened
	the file.
* `TermWidth`: Takes an argument such as TermWidth:80 and adjusts the location
	of the red warning mask to indicate you've typed past the end of a
	terminal width. TermWidth:0 will disable the red mask. This is mostly
	useful to add to the ~/.de/startup file if you have a preference.
* `WarnAlpha`: Takes an argument between 0 and 255 such as WarnAlpha:128 to adjust
	the intensity of the alpha channel used by TermWidth. 255 is fully
	opaque and 1 is mostly transparent. 0 will reset to the default value.
* `Renderer`: Change the renderer used to display the viewport to the renderer
	passed as an argument.
	Valid renderers are:
	*	Renderer:nosyntax - disable syntax highlighting
	*	Renderer:go - use go syntax highlighting
	*	Renderer:php - use php syntax highlighting
	*	Renderer:markdown - use markdown syntax highlighting
	*	Renderer:html - use html syntax highlighting
	*	Renderer:image - directly display PNG or JPEG files in `de`.
			(behaviour on other file types is undefined.)
	*	Renderer:hex - render hex dump of current buffer

## Experimental Commands

These commands are still under development and are likely to have problems.

* `Shell`: Convert the current buffer into an interactive terminal session,
	discarding any changes to the current buffer. This is currently not very
	friendly on CPU resources and doesn't have a renderer which properly
	implements ANSI colour codes yet.
* `Redmine`: Look up an issue on the redmine instance configured in
	 ~/.de/redmine.ini. With no arguments, it will list all projects you
	have access to. With a project as an argument, it will list all issues
	in that project, and with an issue number as an argument it will display
	that issue in the viewport. (This command will probably be removed and
	moved to a different repo, either as a plugin or as a standalone 
	command line command that can be executed from `de` at some point in 
	the future.)

