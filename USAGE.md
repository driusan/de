# de Usage Instructions

## Basic Usage
de is a modal editor that starts in normal mode. The three modes currently implemented are
Normal (biege background), Insert (light green background), and Delete mode (light red/pink
background). The keybindings are inspired by, but not identical to, vi.

In Normal and Delete modes, the movement commands are similar to vi. hjkl move the cursor.
Shift-6 (^) moves the cursor to the start of the line, and Shift-4 ($) moves the cursor to the
end of the line, G moves to end of file (or a line number if you type the line number before hitting
G). w moves the cursor to the next word. In Delete mode it will delete from the current cursor up
to that point, and in Normal mode it will either move to, or select up to that point depending on if 
the ctrl key is pressed. Other movement commands will be implemented, but that's all I've done up to now. 
Unlike in vi, the h and l keys do not stop at line boundaries. p will insert the most recently
delete text, replacing the currently selected text.

Generally, when I find myself typing a keystroke out of muscle-memory enough times as a long
time vi-user, I implement it, or a close enough approximation to it, here. Ranges with a repeat
(ie 3dw) are not yet implemented, but will be eventually.

In Normal mode you can enter Insert mode by pressing 'i' or Delete mode by pressing 'd' and
the background colour should change to indicate the current mode.

In Insert mode, the arrow keys take on the same meaning as hjkl in Normal mode, and Escape will
return to normal mode. Any other key combination that results in a printable unicode character
being sent to de will insert the utf8 encoding of that character at the current location of the
file. In all other modes, the arrow keys scroll the viewport without adjusting the cursor.

In all modes, backspace will delete the currently selected text (or previous character if nothing
is selected) without changing the mode.

Pressing the Escape key will save the current file and exit. *NOTE TO VI USERS: RE-READ THE LAST
SENTENCE*

## Mouse Usage

While the keyboard usage is inspired by vim, the mouse usage is inspired by acme.
The mouse works the same way regardless of keyboard mode.

Clicking anywhere with the left mouse button will move the text cursor or select text.

Clicking with the right mouse button will search for the next instance of the word clicked on
and select the next instance found. If the word is a filename, changes to the current file will be
discarded and the clicked file will be opened in the current window. The keyboard equivalent
for the currently selected text is the slash key (although slash will not open a file.)

Clicking with the middle mouse button will "execute" the word clicked on. (see below.)

Chording will probably eventually work similarly to acme, but isn't yet implemented.

## Executing Words

Words that are selected or clicked on can be "executed" to control the editor, either by
selecting the word and then pressing the Enter key, or by clicking with the middle mouse button.
(When executing with the keyboard, it will first check if the file exists and open it if applicable,
similarly to searching with the mouse.)

If executing a point in a word instead of a selection, that word will be executed.

If the word is an internal editor command, it will perform that command. Otherwise, the shell command
will be executed and the output will replace the currently selected text.

Currently understood commands:
Get (or Discard): Reload the current file from disk and discard changes
Put (or Save): Save the current character buffer to disk, overwriting the existing file.
Exit (or Quit): Quit de, discarding any changes.

To be implemented:
Copy (or Snarf), Cut, Paste,
ACME style tag/Scratch line
