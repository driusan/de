# Contributing to de

If you'd like to contribute to de, I'd love to hear from you, and this document mostly describes
obvious places to get started.

First off, there's the obvious things that every open source project says it needs: bug reports,
bug fixes, documentation, tell your friends, write automated tests, send me feedback, etc. 

On the note of automated tests: I've been a very bad person and haven't been using red,
green, refactor to write this software, despite the fact that Go makes it so easy. The positions
package is a good place to start if you're new to de, because the function names should be pretty
self explanatory about what positions they're supposed to find, they should be unit testable, and
they're probably the highest reward in terms of effort/importance of being well tested. (If you 
have any questions about any of the positions functions, (email me)[mailto:driusan@gmail.com].
I'll try to either respond or update the go docs as appropriate.)

If you're looking for features: renderers for syntax highlighting for different languages that you
use would be welcomed, and (hopefully?) useful to you. I plan on maintaining the Go renderer (for
obvious reasons), the PHP one (so that I can use de at my day job), but I don't use most other
languages often enough to be a good candidate for writing the renderers. I'd gladly accept new
renderer/syntax highlighters from people for other languages.

On that note, the current renderers aren't as DRY as they should be. If you'd like to refactor that
you can create an issue or email me to discuss. (If you don't, I'll get to it eventually.) They
could also use benchmark tests.

Any plugins for new commands that you'd find useful would likely be useful for other people as
well (on top of that, there's the ones that the USAGE.md document says should exist that aren't
yet written.) The worst case scenerio is that your plugin doesn't get accepted, and you can still
put it in your own repository somewhere else and add it to your plugins.go to use it.

For any significant work you should probably create an issue on GitHub with your plan and intended
design and assign it to yourself so the design can be discussed and people don't duplicate work.

I also wouldn't be opposed to anyone doing work that makes progress on GitHub Issue #1 (more
faithful vi key bindings.) The more I use de instead of vi, the more I get used to the idiosyncrasies,
and the less likely I am to remember how it's supposed to work. (Similarly, if you're a long time
acme user, I'm not opposed to any changes that make the mouse usage less surprising to acme users,
either.
