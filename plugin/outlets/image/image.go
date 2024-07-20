package image

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/OutOfBedlam/tine/engine"
	"golang.org/x/image/bmp"
)

func init() {
	engine.RegisterOutlet(&engine.OutletReg{
		Name:    "image",
		Factory: ImageOutlet,
	})
}

func ImageOutlet(ctx *engine.Context) engine.Outlet {
	path := ctx.Config().GetString("path", "")
	fields := ctx.Config().GetStringArray("fields", []string{})
	dstContentType := ctx.Config().GetString("content_type", "image/png")
	jpegQuality := ctx.Config().GetInt("jpeg_quality", 75)
	overwrite := ctx.Config().GetBool("overwrite", false)

	var writer io.Writer
	if w := ctx.Writer(); w != nil {
		writer = w
	}
	dstPattern := path
	if path != "" {
		ext := filepath.Ext(path)
		ext = strings.ToLower(ext)
		switch ext {
		case ".png":
			dstContentType = "image/png"
		case ".jpg", ".jpeg":
			dstContentType = "image/jpeg"
		case ".gif":
			dstContentType = "image/gif"
		case ".bmp":
			dstContentType = "image/bmp"
		}
		ext = filepath.Ext(path)
		base := filepath.Base(path)
		baseWithoutExt := base[:len(base)-len(ext)]
		if overwrite {
			dstPattern = filepath.Join(filepath.Dir(path), baseWithoutExt+"_%s"+ext)
		} else {
			dstPattern = filepath.Join(filepath.Dir(path), baseWithoutExt+"_%s_%d"+ext)
		}
	}
	ctx.SetContentType(dstContentType)

	return &imageOutlet{
		ctx:            ctx,
		dstPath:        path,
		dstContentType: dstContentType,
		srcFields:      fields,
		dstPattern:     dstPattern,
		dstOverwrite:   overwrite,
		dstWriter:      writer,
		jpegQuality:    jpegQuality,
	}
}

type imageOutlet struct {
	ctx            *engine.Context
	dstPath        string
	dstContentType string
	srcFields      []string
	dstSequence    int32
	dstPattern     string
	dstOverwrite   bool
	dstWriter      io.Writer
	// jpeg options
	jpegQuality int
}

func (iout *imageOutlet) Open() error {
	if iout.dstPath == "" {
		return fmt.Errorf("path is not specified")
	}
	if iout.dstContentType == "" {
		return fmt.Errorf("unsupported image format [%s]", filepath.Ext(iout.dstPath))
	}
	return nil
}

func (iout *imageOutlet) Close() error {
	return nil
}

func (iout *imageOutlet) Handle(recs []engine.Record) error {
	for _, rec := range recs {
		if err := iout.encodeRec(rec); err != nil {
			return err
		}
	}
	return nil
}

func (iout *imageOutlet) encodeRec(rec engine.Record) error {
	for _, field := range rec.Fields(iout.srcFields...) {
		if field == nil || field.Type != engine.BINARY {
			continue
		}
		bv := field.Value.(*engine.BinaryValue)
		srcContentType := bv.GetHeader("Content-Type")
		if !strings.HasPrefix(srcContentType, "image/") {
			continue
		}
		if err := iout.writeImageField(field); err != nil {
			return err
		}
	}
	return nil
}

func (iout *imageOutlet) writeImageField(field *engine.Field) error {
	if field.Type != engine.BINARY {
		return fmt.Errorf("field %q is not binary", field.Name)
	}
	bv := field.Value.(*engine.BinaryValue)
	srcContentType := bv.GetHeader("Content-Type")
	if srcContentType == "" {
		return fmt.Errorf("field %q Content-Type is not specified", field.Name)
	}
	srcContentType = strings.Split(srcContentType, ";")[0]
	srcContentType = strings.TrimSpace(srcContentType)
	srcContentType = strings.ToLower(srcContentType)

	var srcImg image.Image
	switch srcContentType {
	default:
		return fmt.Errorf("field %q has unsupported Content-Type %q", field.Name, srcContentType)
	case "image/png":
		if iout.dstContentType != "image/png" {
			if img, err := png.Decode(bytes.NewReader(bv.Data())); err != nil {
				return err
			} else {
				srcImg = img
			}
		}
	case "image/jpeg":
		if iout.dstContentType != "image/jpeg" {
			if img, err := jpeg.Decode(bytes.NewReader(bv.Data())); err != nil {
				return err
			} else {
				srcImg = img
			}
		}
	case "image/gif":
		if iout.dstContentType != "image/gif" {
			if img, err := gif.Decode(bytes.NewReader(bv.Data())); err != nil {
				return err
			} else {
				srcImg = img
			}
		}
	case "image/bmp":
		if iout.dstContentType != "image/bmp" {
			if img, err := bmp.Decode(bytes.NewReader(bv.Data())); err != nil {
				return err
			} else {
				srcImg = img
			}
		}
	case "image/vnd.rgba":
		bin := field.Value.(*engine.BinaryValue)
		strStride := bin.GetHeader("X-RGBA-Stride")
		strRect := bin.GetHeader("X-RGBA-Rectangle")
		stride, err := strconv.ParseInt(strStride, 10, 64)
		if err != nil {
			return fmt.Errorf("field %q has invalid X-RGBA-Stride %q", field.Name, strStride)
		}
		rect, err := ParseImageRectangle(strRect)
		if err != nil {
			return fmt.Errorf("field %q has invalid X-RGBA-Rectangle, %s", field.Name, err.Error())
		}
		srcImg = &image.RGBA{Pix: bin.Data(), Stride: int(stride), Rect: rect}
	}

	var writer io.Writer
	if iout.dstWriter != nil {
		writer = iout.dstWriter
		goto write_image
	}
	if iout.dstOverwrite {
		dstPath := fmt.Sprintf(iout.dstPattern, field.Name)
		if w, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
			return err
		} else {
			writer = w
			defer w.Close()
		}
		iout.ctx.LogDebug("outlets.image", "write_to", dstPath)
	} else {
		dstPath := iout.dstPath
		dstSequence := atomic.LoadInt32(&iout.dstSequence)
	inc_seq:
		if dstSequence > 0 {
			dstPath = fmt.Sprintf(iout.dstPattern, field.Name, dstSequence)
		}
		// check if pe.dstPath file does exists on file system
		if _, err := os.Stat(dstPath); err == nil {
			// File exists
			dstSequence = atomic.AddInt32(&iout.dstSequence, 1)
			goto inc_seq
		} else if os.IsNotExist(err) {
			// File does not exist
		} else {
			return err
		}

		if w, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			return err
		} else {
			writer = w
			defer w.Close()
		}
		iout.ctx.LogDebug("outlets.image", "write_to", dstPath)
	}

write_image:
	if srcImg == nil {
		// no need to convert
		writer.Write(bv.Data())
	} else {
		switch iout.dstContentType {
		case "image/png":
			png.Encode(writer, srcImg)
		case "image/jpeg":
			jpeg.Encode(writer, srcImg, &jpeg.Options{Quality: iout.jpegQuality})
		case "image/gif":
			gif.Encode(writer, srcImg, nil)
		case "image/bmp":
			bmp.Encode(writer, srcImg)
		default:
			iout.ctx.LogError(fmt.Sprintf("unsupported image format [%s]", iout.dstContentType))
		}
	}

	return nil
}
