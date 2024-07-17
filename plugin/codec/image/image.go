package image

import (
	"bytes"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/OutOfBedlam/tine/engine"
)

func init() {
	engine.RegisterEncoder(&engine.EncoderReg{
		Name:        "png",
		Factory:     NewPNGEncoder,
		ContentType: "image/png",
	})
	engine.RegisterEncoder(&engine.EncoderReg{
		Name:        "jpeg",
		Factory:     NewJPEGEncoder,
		ContentType: "image/jpeg",
	})
	engine.RegisterEncoder(&engine.EncoderReg{
		Name:        "gif",
		Factory:     NewGIFEncoder,
		ContentType: "image/gif",
	})
}

func NewPNGEncoder(c engine.EncoderConfig) engine.Encoder {
	return &ImageEncoder{
		EncoderConfig:  c,
		dstContentType: "image/png",
	}
}

func NewJPEGEncoder(c engine.EncoderConfig) engine.Encoder {
	return &ImageEncoder{
		EncoderConfig:  c,
		dstContentType: "image/jpeg",
		jpegQuality:    75,
	}
}

func NewGIFEncoder(c engine.EncoderConfig) engine.Encoder {
	return &ImageEncoder{
		EncoderConfig:  c,
		dstContentType: "image/gif",
	}
}

type ImageEncoder struct {
	engine.EncoderConfig
	dstContentType string
	jpegQuality    int
}

func (pe *ImageEncoder) Encode(recs []engine.Record) error {
	for _, rec := range recs {
		if err := pe.encodeRec(rec); err != nil {
			return err
		}
	}
	return nil
}

func (pe *ImageEncoder) encodeRec(rec engine.Record) error {
	for _, field := range rec.Fields(pe.Fields...) {
		if field == nil {
			continue
		}
		if field.Type == engine.BINARY {
			bv := field.Value.(*engine.BinaryValue)
			srcContentType := bv.GetHeader("Content-Type")
			if srcContentType == "" {
				continue
			}
			srcContentType = strings.Split(srcContentType, ";")[0]
			srcContentType = strings.TrimSpace(srcContentType)
			srcContentType = strings.ToLower(srcContentType)
			switch srcContentType {
			case "image/png":
				switch pe.dstContentType {
				case "image/png":
					pe.Writer.Write(bv.Data())
				case "image/jpeg":
					img, err := png.Decode(bytes.NewReader(bv.Data()))
					if err != nil {
						return err
					}
					jpeg.Encode(pe.Writer, img, &jpeg.Options{Quality: pe.jpegQuality})
				case "image/gif":
					img, err := png.Decode(bytes.NewReader(bv.Data()))
					if err != nil {
						return err
					}
					gif.Encode(pe.Writer, img, nil)
				}
			case "image/jpeg":
				switch pe.dstContentType {
				case "image/png":
					img, err := jpeg.Decode(bytes.NewReader(bv.Data()))
					if err != nil {
						return err
					}
					png.Encode(pe.Writer, img)
				case "image/jpeg":
					pe.Writer.Write(bv.Data())
				case "image/gif":
					img, err := jpeg.Decode(bytes.NewReader(bv.Data()))
					if err != nil {
						return err
					}
					gif.Encode(pe.Writer, img, nil)
				}
			case "image/gif":
				switch pe.dstContentType {
				case "image/png":
					img, err := gif.Decode(bytes.NewReader(bv.Data()))
					if err != nil {
						return err
					}
					png.Encode(pe.Writer, img)
				case "image/jpeg":
					img, err := gif.Decode(bytes.NewReader(bv.Data()))
					if err != nil {
						return err
					}
					jpeg.Encode(pe.Writer, img, &jpeg.Options{Quality: pe.jpegQuality})
				case "image/gif":
					pe.Writer.Write(bv.Data())
				}
			}
		}
	}
	return nil
}
