package chrome

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

func init() {
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

var execPath string
var execPathOnce sync.Once

func (f *chromeSnapFlow) Open() error {
	execPathOnce.Do(func() {
		execPath = FindExecPath()
	})
	if execPath == "" {
		return errors.New("chrome_snap, chrome executable not found")
	}
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

func FindExecPath() string {
	var locations []string
	switch runtime.GOOS {
	case "darwin":
		locations = []string{
			// Mac
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		}
	case "windows":
		locations = []string{
			// Windows
			"chrome",
			"chrome.exe", // in case PATHEXT is misconfigured
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			filepath.Join(os.Getenv("USERPROFILE"), `AppData\Local\Google\Chrome\Application\chrome.exe`),
			filepath.Join(os.Getenv("USERPROFILE"), `AppData\Local\Chromium\Application\chrome.exe`),
		}
	default:
		locations = []string{
			// Unix-like
			"headless_shell",
			"headless-shell",
			"chromium",
			"chromium-browser",
			"google-chrome",
			"google-chrome-stable",
			"google-chrome-beta",
			"google-chrome-unstable",
			"/usr/bin/google-chrome",
			"/usr/local/bin/chrome",
			"/snap/bin/chromium",
			"chrome",
		}
	}

	for _, path := range locations {
		found, err := exec.LookPath(path)
		if err == nil {
			return found
		}
	}
	return ""
}
