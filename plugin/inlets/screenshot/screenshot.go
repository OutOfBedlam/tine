package screenshot

import (
	"bytes"
	"fmt"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"sync/atomic"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/kbinani/screenshot"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "screenshot",
		Factory: ScreenshotInlet,
	})
}

func ScreenshotInlet(ctx *engine.Context) engine.Inlet {
	interval := ctx.Config().GetDuration("interval", 10*time.Second)
	count := ctx.Config().GetInt("count", 0)
	displays := ctx.Config().GetIntArray("displays", nil)
	format := ctx.Config().GetString("format", "rgba")

	return &screenshotInlet{
		ctx:      ctx,
		interval: interval,
		count:    count,
		displays: displays,
		format:   format,
	}
}

type screenshotInlet struct {
	ctx      *engine.Context
	interval time.Duration
	count    int
	runCount int32
	displays []int
	format   string
}

var _ = engine.PullInlet((*screenshotInlet)(nil))

func (si *screenshotInlet) Open() error {
	return nil
}

func (si *screenshotInlet) Close() error {
	return nil
}

func (si *screenshotInlet) Interval() time.Duration {
	return si.interval
}

func (si *screenshotInlet) Pull() ([]engine.Record, error) {
	runCount := atomic.AddInt32(&si.runCount, 1)
	if si.count > 0 && int(runCount) > si.count {
		return nil, io.EOF
	}

	rec := engine.NewRecord()

	// Capture each display
	numOfDisp := screenshot.NumActiveDisplays()
	for disp := 0; disp < numOfDisp; disp++ {
		if len(si.displays) > 0 {
			found := false
			for _, d := range si.displays {
				if d == disp {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		bounds := screenshot.GetDisplayBounds(disp)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			return nil, err
		}
		var bin *engine.BinaryValue
		switch si.format {
		default: // "rgba"
			bin = engine.NewBinaryValue(img.Pix)
			bin.SetHeader("Content-Type", "image/vnd.rgba")
			bin.SetHeader("X-RGBA-Stride", fmt.Sprintf("%d", img.Stride))
			bin.SetHeader("X-RGBA-Rectangle", fmt.Sprintf("%d,%d,%d,%d", img.Rect.Min.X, img.Rect.Min.Y, img.Rect.Max.X, img.Rect.Max.Y))
		case "png":
			buf := &bytes.Buffer{}
			if err := png.Encode(buf, img); err != nil {
				return nil, err
			}
			bin = engine.NewBinaryValue(buf.Bytes())
			bin.SetHeader("Content-Type", "image/png")
		case "jpeg":
			buf := &bytes.Buffer{}
			if err := jpeg.Encode(buf, img, nil); err != nil {
				return nil, err
			}
			bin = engine.NewBinaryValue(buf.Bytes())
			bin.SetHeader("Content-Type", "image/jpeg")
		case "gif":
			buf := &bytes.Buffer{}
			if err := gif.Encode(buf, img, nil); err != nil {
				return nil, err
			}
			bin = engine.NewBinaryValue(buf.Bytes())
			bin.SetHeader("Content-Type", "image/gif")
		}
		if bin == nil {
			continue
		}
		bin.SetHeader("X-Screenshot-Display", fmt.Sprintf("%d", disp))
		rec = rec.Append(
			engine.NewBinaryField(fmt.Sprintf("display_%d", disp), bin),
		)
	}

	ret := []engine.Record{rec}
	if si.count > 0 && int(runCount) >= si.count {
		return ret, io.EOF
	}
	return ret, nil
}
