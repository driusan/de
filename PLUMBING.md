# Plumbing

(If you're unfamiliar with the Plan 9 concept of "plumbing", you may want
to start by reading [the paper](http://doc.cat-v.org/plan_9/4th_edition/papers/plumb))

de now supports directly using the plan9ports plumber for interpreting
interactions, if it's available, with a couple caveats.

1. Since de is a single-buffer editor, we need another process (the deplumber)
   to handle the incoming messages. (If de handled the edit port itself either 
   either every window would process each edit message and the windows
   would multiply like rabbits, or the plumbing would stop working after the
   first window closed.)
2. The default p9p plumbing rules don't plumb directories to $editor. The file
   plumbing.sample is a sample config that you can put in $HOME/lib/plumbing
   in order to use deplumber to listen on the edit port, and also plumb
   directories to edit.

So to use de with plumbing the steps are:
1. Run "plumber" (from p9p)
2. Run "deplumber" (from de, you may need to `go get github.com/driusan/de/...` first) in the background)
3. Run de, or plumb a message some other way to test it.

You can plumb from de either by hitting the Enter key or right clicking
somewhere. If the message couldn't be successfully plumbed or deplumber isn't
running, de will fall back on the old behaviour (find next for right-click, and
execute for enter.)

de is a text editor, not a window manager. If you're running under X11, you may
want to use a tiling window manager in order to have your windows managed in a
way that makes de more closely resemble acme. Otherwise, de doesn't pretend
that it knows how you like your windows arranged better than you and your window
manager.

The integration of plumbing directly into de is relatively new, so if you have
any problems, please file a bug report. 