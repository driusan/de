package viewer

// 82 characters, for testing TermWidth:80 and WarnAlpha:16
// (0-9 8 times, plus two slashes for the comment.)
//12345678901234567890123456789012345678901234567890123456789012345678901234567890
import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/driusan/de/demodel"
	"github.com/driusan/de/renderer"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/math/fixed"
)

type Viewport struct {
	demodel.Map
	demodel.Renderer
	Location image.Point

	kbLocked bool
	Window   screen.Window

	// the number of characters after which the viewport should put
	// a red warning mask. Usually 80 chars
	termwidth int
	warnalpha uint8
}

type RequestRerender struct{}

func (v *Viewport) Rerender() {
	v.Window.Send(RequestRerender{})

}
func (v *Viewport) GetKeyboardMode() demodel.Map {
	return v.Map
}

var ErrKBLocked = errors.New("Keyboard mode is locked")
var ErrInvalidViewport = errors.New("Invalid viewport")

func (v *Viewport) ResetLocation() error {
	if v == nil {
		return ErrInvalidViewport
	}
	v.Location.X = 0
	v.Location.Y = 0
	return nil
}

// Sets the keyboard mode in a way that future requests to SetKeyboardMode
// will fail.
func (v *Viewport) LockKeyboardMode(m demodel.Map) error {
	if v.kbLocked {
		return ErrKBLocked
	}
	v.kbLocked = true
	v.Map = m
	return nil
}

// Unlocks a locked keyboard mode. The locked mode must be passed as a
// parameter to ensure that it's the same caller that locked it.
// It's not secure, but it prevents plugins from accidentally unlocking
// someone else's locked keyboard.
// Returns ErrKBLocked if d doesn't equal the mode that it's locked to.
func (v *Viewport) UnlockKeyboardMode(d demodel.Map) error {
	if v.Map == d {
		v.kbLocked = false
		return nil
	}
	return ErrKBLocked
}
func (v *Viewport) SetKeyboardMode(m demodel.Map) error {
	if v.kbLocked {
		return ErrKBLocked
	}
	v.Map = m
	//fmt.Print("Set keyboard mode!")
	return nil
}

func (v *Viewport) GetRenderer() demodel.Renderer {
	return v.Renderer
}

func (v *Viewport) SetRenderer(r demodel.Renderer) error {
	v.Renderer = r
	return nil
}

func (v *Viewport) SetOption(opt string, val interface{}) error {
	switch opt {
	case "TermWidth":
		if i, ok := val.(int); ok {
			v.termwidth = i
			return nil
		}
		return fmt.Errorf("TermWidth must be an integer")
	case "WarnAlpha":
		if i, ok := val.(uint8); ok {
			v.warnalpha = i
			return nil
		}
		return fmt.Errorf("WarnAlpha must be in the range 0-255")

	default:
		return demodel.ErrUnsupportedOption
	}
}

func mult(x, y fixed.Int26_6) fixed.Int26_6 {
	// multiplying two fixed-point precision values requires shifting
	// the decimal 6 places to get it back to the right place. This will
	// overflow if either x or y are greater than (1 << (26-6)), but if
	// anyone is using a window that width they'll probably have memory
	// problems long before getting into issues with this math.
	return x * y >> 6
}
func (v *Viewport) RenderInto(dst draw.Image, buffer *demodel.CharBuffer, viewport image.Rectangle) error {
	err := v.Renderer.RenderInto(dst, buffer, viewport)
	if err != nil || v.termwidth == 0 {
		return err
	}

	fontAdvance, ok := renderer.MonoFontFace.GlyphAdvance('x')
	if !ok {
		// if the actual renderer succeeded, don't be overly concerned about errors
		// drawing the red warning..
		return nil
	}

	bounds := dst.Bounds()

	var r image.Rectangle
	if v.termwidth > 0 {
		// a positive value means draw a mask after that many characters
		r = image.Rectangle{
			image.Point{
				mult(fontAdvance, fixed.I(v.termwidth)).Ceil() - viewport.Min.X,
				bounds.Min.Y,
			},
			image.Point{bounds.Max.X, bounds.Max.Y},
		}

	} else {
		// A negative value means to draw the mask over the non-overflow
		// section.
		r = image.Rectangle{
			image.ZP,
			image.Point{
				mult(fontAdvance, fixed.I(-v.termwidth)).Ceil(),
				bounds.Max.Y,
			},
		}
	}

	warnalpha := v.warnalpha
	if warnalpha == 0 {
		// 8/255 = 3.125% alpha mask.
		// It's subtle enough to not be too distracting while
		// still being noticable if you look for it.
		// Anyone who has problems with it can customize with
		// WarnAlpha:128 (or some other mask between 0-255)
		warnalpha = 8
	}

	draw.DrawMask(dst,
		r,
		&image.Uniform{color.RGBA{255, 0, 0, 255}},
		image.ZP,
		&image.Uniform{color.RGBA{0, 0, 0, warnalpha}},
		image.ZP,
		draw.Over,
	)

	return nil
}
