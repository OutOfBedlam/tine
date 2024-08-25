package qrcode

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/OutOfBedlam/tine/engine"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

func init() {
	engine.RegisterFlow(&engine.FlowReg{
		Name:    "qrcode",
		Factory: QRCodeFlow,
	})
}

func QRCodeFlow(ctx *engine.Context) engine.Flow {
	return &qrCodeFlow{ctx: ctx}
}

type qrCodeFlow struct {
	ctx         *engine.Context
	inputField  string
	outputField string
	options     []standard.ImageOption
}

var _ = engine.Flow((*qrCodeFlow)(nil))

func (qf *qrCodeFlow) Open() error {
	cfg := qf.ctx.Config()
	qf.inputField = cfg.GetString("input_field", "")
	qf.outputField = cfg.GetString("output_field", "qrcode")
	if qf.inputField == "" {
		return fmt.Errorf("flows.qrcode input_field is required")
	}

	width := uint8(cfg.GetInt("width", 21))
	bgColor := cfg.GetString("background_color", "#ffffff")
	fgColor := cfg.GetString("foreground_color", "#000000")

	qf.options = []standard.ImageOption{
		standard.WithQRWidth(width),
		standard.WithBgColorRGBHex(bgColor),
		standard.WithFgColorRGBHex(fgColor),
		standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
	}
	if logo := cfg.GetString("logo", ""); logo != "" {
		if filepath.Ext(logo) == ".png" {
			qf.options = append(qf.options, standard.WithLogoImageFilePNG(logo))
		} else if filepath.Ext(logo) == ".jpeg" {
			qf.options = append(qf.options, standard.WithLogoImageFileJPEG(logo))
		}
	}
	if halftone := cfg.GetString("halftone", ""); halftone != "" {
		qf.options = append(qf.options, standard.WithHalftone(halftone))
	}
	if cfg.GetBool("background_transparent", false) {
		qf.options = append(qf.options, standard.WithBgTransparent())
	}
	return nil
}

func (qf *qrCodeFlow) Close() error     { return nil }
func (qf *qrCodeFlow) Parallelism() int { return 1 }

func (qf *qrCodeFlow) Process(records []engine.Record, nextFunc engine.FlowNextFunc) {
	for i, rec := range records {
		if in := rec.Field(qf.inputField); in != nil {
			if out, err := qf.gen(in); err != nil {
				qf.ctx.LogError("failed to generate qrcode", "error", err)
				nextFunc(nil, err)
			} else {
				records[i] = rec.AppendOrReplace(out)
			}
		}
	}
	nextFunc(records, nil)
}

func (qf *qrCodeFlow) gen(in *engine.Field) (*engine.Field, error) {
	text, ok := in.Value.String()
	if !ok {
		return nil, fmt.Errorf("field %s is not a string", qf.outputField)
	}
	qrc, err := qrcode.New(text)
	if err != nil {
		return nil, err
	}

	buff := &bytes.Buffer{}
	w0 := standard.NewWithWriter(engine.NopCloser(buff), qf.options...)
	err = qrc.Save(w0)
	if err != nil {
		return nil, err
	}
	out := engine.NewField(qf.outputField, buff.Bytes())
	out.Tags.Set("Content-Type", engine.NewValue("image/png"))
	return out, nil
}
