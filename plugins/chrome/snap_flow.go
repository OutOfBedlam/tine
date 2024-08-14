package chrome

import (
	"context"
	"runtime"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

func init() {
	if runtime.GOOS != "linux" {
		return
	}
	engine.RegisterFlow(&engine.FlowReg{
		Name:    "chrome_snap",
		Factory: ChromeSnapFlow,
	})
}

func ChromeSnapFlow(ctx *engine.Context) engine.Flow {
	conf := ctx.Config()
	parallelism := conf.GetInt("parallelism", 1)
	timeout := conf.GetDuration("timeout", 10)
	urlField := conf.GetString("url_field", "url")
	outField := conf.GetString("out_field", "chrome_snap")
	userAgent := conf.GetString("user_agent", "")
	fullPage := conf.GetInt("full_page", 0)
	userDevice := conf.GetString("device", "")
	return &chromeSnapFlow{
		ctx:             ctx,
		parallelism:     parallelism,
		urlField:        urlField,
		outField:        outField,
		timeout:         timeout,
		userAgent:       userAgent,
		userDevice:      userDevice,
		fullPageQuality: fullPage,
	}
}

type chromeSnapFlow struct {
	ctx             *engine.Context
	parallelism     int
	urlField        string
	outField        string
	timeout         time.Duration
	userAgent       string
	userDevice      string
	fullPageQuality int
}

func (f *chromeSnapFlow) Open() error {
	return nil
}

func (f *chromeSnapFlow) Close() error {
	return nil
}

func (f *chromeSnapFlow) Parallelism() int {
	return f.parallelism
}

// This requires install google chrome.
//
// wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
// sudo apt -y install ./google-chrome-stable_current_amd64.deb
func (f *chromeSnapFlow) Process(recs []engine.Record, nextFunc engine.FlowNextFunc) {
	for _, rec := range recs {
		urlField := rec.Field(f.urlField)
		if urlField == nil {
			nextFunc([]engine.Record{rec}, nil)
			continue
		}
		url, ok := urlField.Value.String()
		if !ok {
			nextFunc([]engine.Record{rec}, nil)
			continue
		}
		ctx, cancel := chromedp.NewContext(f.ctx)
		defer cancel()

		ctx, cancel = context.WithTimeout(ctx, f.timeout)
		defer cancel()

		var buf []byte

		actions := []chromedp.Action{
			chromedp.Navigate(url),
		}
		if f.userAgent != "" {
			actions = append(actions, emulation.SetUserAgentOverride(f.userAgent))
		}
		if f.userDevice != "" {
			if dev, ok := UserDevice(f.userDevice); ok {
				actions = append(actions, chromedp.Emulate(dev))
			}
		}
		if f.fullPageQuality > 0 {
			actions = append(actions, chromedp.FullScreenshot(&buf, f.fullPageQuality))
		} else {
			actions = append(actions, chromedp.CaptureScreenshot(&buf))
		}

		if err := chromedp.Run(ctx, actions...); err != nil {
			f.ctx.LogError("chrome_snap", err)
			nextFunc([]engine.Record{rec}, err)
		} else {
			imgField := engine.NewField(f.outField, buf)
			imgField.Tags[engine.CanonicalTagKey("Content-Type")] = engine.NewValue("image/png")
			rec = rec.AppendOrReplace(imgField)
			nextFunc([]engine.Record{rec}, nil)
		}
	}
}
