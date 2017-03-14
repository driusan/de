package demodel

import (
	"errors"
	"image"
	"image/draw"
)

var ErrUnsupportedOption = errors.New("Unsupported Option")

// A Renderer takes a character buffer and renders it to an image to be displayed in a viewport.
type Renderer interface {
	// Determines whether this renderer knows how to render this character buffer.
	// The most recently registered plugin that can handle it wins (because otherwise
	// everything would use the default renderer, which claims it can render anything.)
	CanRender(*CharBuffer) bool

	// Given a character buffer, and a clipping region (viewport) to be displayed, renders an image that
	// should be shown on the screen, a rectangle determining what the size of the *entire* buffer rendered
	// would be (even if it didn't render it), an ImageMap that, at least for the portion rendered, can
	// be used to determine what any pixel represents, and an error (hopefully nil.)
	RenderInto(dst draw.Image, buffer *CharBuffer, viewport image.Rectangle) error

	// Returns the bounds that would render the entire buffer, if the viewport were big enough.
	Bounds(buffer *CharBuffer) image.Rectangle

	// Returns an ImageMap covering the entire bounds
	GetImageMap(buffer *CharBuffer, viewport image.Rectangle) ImageMap
	// This being in the interface is a temporary hack until the renderers are refactored to share
	// more code. It requests that any cache of the image size be invalidated, because the DPI
	// changed. It doesn't belong here, but it's got nowhere else to go right now.
	InvalidateCache()
}

// A Viewport represents the state of the window being rendered.
type Viewport interface {
	Renderer
	// Returns the current KBMap mode that the viewport is in.
	GetKeyboardMode() Map
	// Requests that the KBMap be changed to a new mode for this viewport.
	SetKeyboardMode(Map) error
	// Requests that the KBMap be changed to a new mode, and further changes
	// be disallowed until it's explicitly unlocked. This is mostly for plugins
	// such as Shell
	LockKeyboardMode(Map) error
	// Informs the viewport that the keyboard map should be unlocked from Map.
	// The current map must be passed as a parameter, mostly to prevent accidentally
	//unlocking someone else's lock, not for security reasons.
	UnlockKeyboardMode(Map) error

	// Registers a channel to be notified after a mouse event is processed.
	RegisterMouseListener(chan interface{})
	DeregisterMouseListener(chan interface{})

	// Request a change of the renderer to be used to render the main contents
	// of this viewport
	SetRenderer(Renderer) error

	// Request that the viewport be rerendered.
	Rerender()

	ResetLocation() error

	SetOption(option string, value interface{}) error
}

// An ImageMap represents a way to convert points in an image into indexes into a
// character buffer and vice versa. They generally, internally, have a pointer to
// a char buffer that they're rendering and an image that they're mapping to.
type ImageMap interface {
	// At returns the index into the charbuffer that this image map is representing
	// at point x,y for the image which is being mapped.
	At(x, y int) (uint, error)

	// Get returns the bounding rectangle for the glyph at index idx of the CharBuffer
	// being mapped in the image being mapped to.
	Get(idx uint) (image.Rectangle, error)
}
