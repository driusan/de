# de Usage Instructions

## Basic Usage

de is a modal editor that starts in normal mode. The three modes
currently implemented are Normal (biege background), Insert (light green
background), and Delete mode (light red/pink background). The
keybindings are inspired by, but not identical to, vi.

### Normal Mode

Movements can be repeated multiple times by prefixing the command with a
number (ie `3w` to go forward 3 words)

In Normal mode, typing a movement command moves the cursor to that
position, like vi. Unlike vi, holding CTRL and typing a movement command
instead expands the selection by that much movement similar to visual
mode in vi.

-   `h`: move left by *count* columns
-   `j`: move down by *count* lines
-   `k`: move up by *count *lines
-   `l`: move right by *count* columns

-   `w`: move forward *count* words
-   `b`: move back *count* words
-   `$`: move to end of line
-   `^`: move to start of line
-   `G`: move to line, or last line without a prefix

Add new text by entering Insert mode with `i`, `a` or `o`.

-   `i`: insert before cursor
-   'I': insert at start of line (after any whitespace characters.)
-   `a`: append after cursor
-   'A': append at end of the current line.
-   `o`: open a new line after the current line.

Delete existing text can either happen with shorthands like Backspace or
`x`, or with full Delete mode.

-   `d`: enter Delete mode
-   `x`: delete currently selected text, or *count* characters forward
-   Backspace: deletecurrently selected text, or previous character if
    nothing is selected

Some other common commands:

-   `p`: **Paste**
-   `u`: **Undo**
-   `J`: **Join** the selected lines together into a single line.
-   `;`: give keyboard focus to tagline (see below for tagline usage)
-   `:`: give keyboard focus to end of tagline (see below)
-   Enter: open or perform action on currently selected text, or current
    word if nothing is selected
-   `/`: move to next occurence of currently selected text, or current
    word if nothing is selected

-   Left: scroll viewport left
-   Up: scroll viewport up
-   Down: scroll viewport down
-   Right: scroll viewport right

-   Esc: **SaveOrExit**: **Save** if buffer has been modified, **Exit** if buffer has not been
    modified (so hitting Escape twice will **Save** and **Exit**)

### Insert Mode

In Insert mode, the arrow keys take on the same meaning as hjkl in
Normal mode, and Escape will return to normal mode. Any other key
combination that results in a printable unicode character being sent to
de will insert the utf8 encoding of that character at the current
location of the file.

-   Left: move cursor left one char
-   Up: move cursor up one line
-   Down: move cursor down one line
-   Right: move cursor right one char

-   Escape: exit Insert mode to Normal mode

### Delete Mode

Delete mode is similar to Normal mode, except the selected text plus the
movement from the beginning or end will be deleted (depending on if it's
a forward or backwards movement command.)

The most recently deleted text will be put in the snarf buffer (or
"clipboard", if you prefer), overwriting what was there previously.

-   `h`: delete left *count* chars
-   `j`: delete down *count* char
-   `k`: delete up *count* chars
-   `l`: delete right *count* char

-   `x`: delete selected text, or current char if nothing selected
-   `w`: delete to next *count* words
-   `b`: delete to previous *count* words start
-   `$`: delete to end of line
-   `^`: delete to start of line
-   `G`: delete to line, or end of file with no prefix

-   `d`: delete *count* entire lines
-   Backspace: delete selected text, or previous char if nothing selected

-   Left: scroll viewport left
-   Up: scroll viewport up
-   Down: scroll viewport down
-   Right: scroll viewport right

-   Escape: exit Delete mode to Normal mode (without deleting anything)

## Mouse Usage

While the keyboard usage is inspired by vim, the mouse usage is inspired by acme.
The mouse works the same way regardless of keyboard mode.

Clicking anywhere with the left mouse button will move the text cursor or select text.

Clicking with the right mouse button will search for the next instance of the word clicked on
and select the next instance found. If the word is a filename, changes to the current file will be
discarded and the clicked file will be opened in the current window. The keyboard equivalent
for the currently selected text is the slash key (although slash will not open a file.)

Clicking with the middle mouse button will "execute" the word clicked on. (see below.)

Chording will probably eventually work similarly to acme, but isn't yet implemented, since my
laptop doesn't have a three button mouse.

## Executing Words

Words that are selected or clicked on can be "executed" to control the editor, either by
selecting the word and then pressing the Enter key, or by clicking with the middle mouse button.
(When executing with the keyboard, it will first check if the file exists and open it if applicable,
similarly to searching with the mouse.)

For details about all built in commands, see COMMANDS.md.

If executing a point in a word instead of a selection, that word will be executed.

If the word is an internal editor command, it will perform that command. Otherwise, it will be
executed as a shell command and the output to stdout from that command will replace the currently
selected text.

The top of the window has a tag line that you can use as a scratch space for writing commands
without affecting the content of the current file. Typing in the tag line always acts like it's
in Insert mode, with the exception that the return key works as if it were in Normal mode and
either opens or executes the current word or selected text. You can access the tagline either by pointing
to it (focus follows pointer, you don't need to click) or typing the ; key. ; will just give focus
to the tag line, and : will give it focus as well as moving the cursor to the end and appending a
space. This means that you can type, for instance, :Save<Enter> to save the current file.

The tagline doubles as a Stderr channel for de, and some commands will use it as an area to provide
status or error messages without affecting your editing buffer. (However the errors are still just
regular text that can be executed, edited, or deleted when you're done with them.)

Note to Plan 9 users: you can *not* change the filename by editing it in the tagline and putting. The
filename there is only for reference, and updated when a new file is opened if the current filename
happens to be a prefix. There is currently no way to change the filename and save the file to another
name.

When the word (or selection) isn't an internal plugin command (generally commands with a capital first
letter by convention, although that's not enforced), de will try to execute the shell command selected
and pass the selected text (or the whole file if nothing is selected) to the processes's STDIN.

In the general case, if the command executes successfully, de will take the stdout of the command,
insert it *after* the currently selected text, and then select the newly inserted text. This is generally
safe, because if something went wrong you can just hit backspace or x to delete it.

The behaviour of what to do with the output can be modified by prefixing the command with either < or |.
< generally means "Replace the whole buffer with this command's output" and "|" means "Replace the
selected text with the output of filtering it through the command." Either one is most useful in the
tagline, where you can put arbitrary sed commands, or gofmt, or any arbitrary script that you wrote
which reads from stdin and outputs to stdout. Prefixing the whole thing with ! will tell de to ignore
the return code, and insert the output even if the process reports an error.

Finally, the scripts directory contains some wrapper scripts that may be useful to put in your path
and execute from the tagline.

## Startup Commands

At startup, de will run any Commands in `~/.de/startup`. At the moment, this is
probably only useful for setting **TermWidth**.
