package args

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "args",
		Factory: ArgsInlet,
	})
}

func ArgsInlet(ctx *engine.Context) engine.Inlet {
	return &argsInlet{ctx: ctx}
}

type argsInlet struct {
	ctx        *engine.Context
	httpClient *http.Client
}

var _ = engine.PushInlet((*argsInlet)(nil))

func (ai *argsInlet) Open() error {
	return nil
}

func (ai *argsInlet) Close() error {
	return nil
}

func (ai *argsInlet) Push(cb func([]engine.Record, error)) {
	foundDoubleDash := false
	keys := []string{}
	vals := []string{}
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "--" {
			foundDoubleDash = true
			continue
		}
		if !foundDoubleDash {
			continue
		}
		str := os.Args[i]
		for strings.HasPrefix(str, "-") {
			str = strings.TrimPrefix(str, "-")
		}
		toks := strings.SplitN(str, "=", 2)
		if len(toks) != 2 && len(os.Args) > i+1 {
			keys = append(keys, str)
			vals = append(vals, os.Args[i+1])
			i++
		} else {
			keys = append(keys, toks[0])
			vals = append(vals, toks[1])
		}
	}
	rec := engine.NewRecord()
	for i := 0; i < len(keys); i++ {
		name := keys[i]
		value := vals[i]
		if strings.HasPrefix(value, "binary+file://") {
			value = value[14:]
			body, contentType, err := ai.fetchFile(value)
			if err != nil {
				cb(nil, err)
				return
			}
			bv := engine.NewBinaryField(name, body)
			bv.Tags.Set(engine.CanonicalTagKey("Content-Type"), engine.NewValue(contentType))
			rec = rec.Append(bv)
		} else if strings.HasPrefix(value, "binary+http://") || strings.HasPrefix(value, "binary+https://") {
			value = value[7:]
			body, contentType, err := ai.fetchHttp(value)
			if err != nil {
				cb(nil, err)
				return
			}
			bv := engine.NewBinaryField(name, body)
			bv.Tags.Set(engine.CanonicalTagKey("Content-Type"), engine.NewValue(contentType))
			rec = rec.Append(bv)
		} else if strings.HasPrefix(value, "base64+file://") {
			value = value[14:]
			body, err := os.ReadFile(value)
			if err != nil {
				cb(nil, err)
				return
			}
			base64ed := base64.StdEncoding.EncodeToString(body)
			rec = rec.Append(engine.NewStringField(name, base64ed))
		} else if strings.HasPrefix(value, "base64+http://") || strings.HasPrefix(value, "base64+https://") {
			value = value[7:]
			body, _, err := ai.fetchHttp(value)
			if err != nil {
				cb(nil, err)
				return

			}
			base64ed := base64.StdEncoding.EncodeToString(body)
			rec = rec.Append(engine.NewStringField(name, base64ed))
		} else {
			rec = rec.Append(engine.NewStringField(name, value))
		}
	}
	cb([]engine.Record{rec}, nil)
}

func (ai *argsInlet) fetchFile(path string) ([]byte, string, error) {
	body, err := os.ReadFile(path)
	contentType := ""
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		contentType = "image/png"
	case ".jpg":
		contentType = "image/jpeg"
	case ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	default:
		contentType = "application/octet-stream"
	}
	return body, contentType, err
}

func (ai *argsInlet) fetchHttp(addr string) ([]byte, string, error) {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		return nil, "", fmt.Errorf("unsupported protocol: %s", addr)
	}
	if ai.httpClient == nil {
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}
		ai.httpClient = &http.Client{
			Transport: transport,
		}
	}
	rsp, err := ai.httpClient.Get(addr)
	if err != nil {
		return nil, "", err
	}
	defer rsp.Body.Close()
	if body, err := io.ReadAll(rsp.Body); err != nil {
		return nil, "", err
	} else {
		return body, rsp.Header.Get("Content-Type"), nil
	}
}
