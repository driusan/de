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
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type LineNumberMode int

const (
	NoLineNumbers = LineNumberMode(iota)
	RelativeLineNumbers
	AbsoluteLineNumbers
)

// BackgroundMode represents whether or not the viewport background should
// change colour depending on the mode. It can be disabled by setting the
// viewport mode to StableBackground.
type BackgroundMode uint8

const (
	// Background is context-sensitive
	DefaultBackground = BackgroundMode(iota)
	// background does not change colour
	StableBackground
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

	lineNumberMode LineNumberMode
	BackgroundMode BackgroundMode
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
	return nil
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
	case "LineNumbers":
		if b, ok := val.(LineNumberMode); ok {
			v.lineNumberMode = b
			return nil
		}
		return fmt.Errorf("Invalid parameter for LineNumbers command")
	case "RotateLineNumbers":
		switch v.lineNumberMode {
		case NoLineNumbers:
			v.lineNumberMode = RelativeLineNumbers
		case RelativeLineNumbers:
			v.lineNumberMode = AbsoluteLineNumbers
		default:
			v.lineNumberMode = NoLineNumbers
		}
		return nil
	case "BackgroundMode":
		if m, ok := val.(BackgroundMode); ok {
			v.BackgroundMode = m
		} else {
			v.BackgroundMode = DefaultBackground
		}
		return nil
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
	var lineNumberOffset fixed.Int26_6
	switch v.lineNumberMode {
	case RelativeLineNumbers, AbsoluteLineNumbers:
		lineNumberOffset = renderer.MonoFontAdvance * 6
	}
	rImg := dst.(*image.RGBA)
	bounds := dst.Bounds()
	err := v.Renderer.RenderInto(
		rImg.SubImage(
			image.Rectangle{
				image.Point{lineNumberOffset.Ceil() + bounds.Min.X, bounds.Min.Y},
				bounds.Max,
			}).(*image.RGBA),
		buffer,
		viewport,
	)
	if err != nil || v.termwidth == 0 {
		return err
	}

	var r image.Rectangle
	if v.termwidth > 0 {
		// a positive value means draw a mask after that many characters
		r = image.Rectangle{
			image.Point{
				(mult(renderer.MonoFontAdvance, fixed.I(v.termwidth)) + lineNumberOffset).Ceil() - viewport.Min.X,
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
				(mult(renderer.MonoFontAdvance, fixed.I(-v.termwidth)) + lineNumberOffset).Ceil(),
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

	// Print the line numbers. Neither the relative nor absolute line numbers
	// code are very efficient, but they work and efficiency appears good
	// enough for average source code size files. For any files where this
	// is a problem, LineNumbers:off doesn't incur the overhead.
	switch v.lineNumberMode {
	case RelativeLineNumbers:
		// LineNumbers:relative
		im := v.Renderer.GetImageMap(buffer, viewport)
		var p image.Rectangle
		if im == nil {
			return fmt.Errorf("No imagemap for viewport")
		}

		if buffer.Dot.Start >= uint(len(buffer.Buffer)) {
			p, _ = im.Get(uint(len(buffer.Buffer)) - 1)
		} else {
			p, _ = im.Get(buffer.Dot.Start)
		}

		writer := font.Drawer{
			Dst:  dst,
			Src:  &image.Uniform{color.Black},
			Face: renderer.MonoFontFace,
			Dot:  fixed.P(0, p.Max.Y+renderer.MonoFontAscent.Floor()),
		}

		i := int64(0)
		// negative relative line numbers
		for y := p.Max.Y; y >= bounds.Min.Y; y -= (p.Max.Y - p.Min.Y) - 1 {
			writer.Dot.Y = fixed.I(y - viewport.Min.Y + renderer.MonoFontAscent.Floor())
			writer.Dot.X = 0

			writer.DrawString(fmt.Sprintf("%d", i))
			i--
		}

		// positive relative line numbers
		i = 0

		if buffer.Dot.End >= uint(len(buffer.Buffer)) {
			p, _ = im.Get(uint(len(buffer.Buffer)) - 1)
		} else {
			p, _ = im.Get(buffer.Dot.End)
		}
		imend, _ := im.Get(uint(len(buffer.Buffer)) - 1)
		imend.Max.Y += renderer.MonoFontHeight.Ceil()

		writer.Dot.Y = fixed.I(p.Max.Y - viewport.Min.Y + renderer.MonoFontAscent.Floor())
		writer.Dot.X = 0

		// The writer's dot maintains itself in reference to the image being
		// drawn into, this maintains the position in absolute terms of the
		// buffer, to make it easier to tell if we've gotten past the end
		// of the image.
		absDot := fixed.I(p.Max.Y)
		for {
			writer.DrawString(fmt.Sprintf("%d", i))
			writer.Dot.Y += renderer.MonoFontHeight + 1
			writer.Dot.X = 0
			i++

			absDot += renderer.MonoFontHeight + 1
			if absDot.Ceil() >= viewport.Max.Y || absDot.Ceil() >= imend.Max.Y {
				break
			}
		}
	case AbsoluteLineNumbers:
		// LineNumbers:absolute
		b := dst.Bounds()

		im := v.Renderer.GetImageMap(buffer, viewport)
		if im == nil {
			return fmt.Errorf("No imagemap for viewport")
		}

		startChar, err := im.At(0, viewport.Min.Y)
		if err != nil {
			return err
		}
		p, err := im.Get(startChar)
		if err != nil {
			return err
		}

		writer := font.Drawer{
			Dst:  dst,
			Src:  &image.Uniform{color.Black},
			Face: renderer.MonoFontFace,
			Dot:  fixed.P(0, b.Min.Y+renderer.MonoFontAscent.Floor()-(viewport.Min.Y-p.Min.Y)),
		}

		// This is a horrible, inefficient way to get the line number
		// that we start at, but it seems to be more accurate than the
		// alternatives that I've tried, which mostly seem to have
		// problems with rounding error caused by conversions between int
		// and fixed.Int26_6 in various places. The rounding errors
		// compound with absolute line number mode depending on how far
		// down you've scrolled, making them inaccurate, but aren't a
		// problem with relative line number mode, because regardless of
		// how far down you've scrolled it starts at 0.
		lineNo := uint(1)

		if buffer.Buffer[startChar] == '\n' {
			lineNo--
		}

		for i, r := range buffer.Buffer {
			switch r {
			case '\n':
				lineNo++
			}
			if uint(i) >= startChar {
				break
			}
		}

		imend, _ := im.Get(uint(len(buffer.Buffer)) - 1)

		for {
			writer.DrawString(fmt.Sprintf("%d", lineNo))
			writer.Dot.Y += renderer.MonoFontHeight + 1
			writer.Dot.X = 0
			if writer.Dot.Y.Ceil() > b.Max.Y || viewport.Min.Y+writer.Dot.Y.Floor()-b.Min.Y >= imend.Max.Y {
				break
			}
			lineNo++
		}
	}
	return nil
}
