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
	dstPathField := ctx.Config().GetString("path_field", "")
	fields := ctx.Config().GetStringSlice("image_fields", []string{})
	jpegQuality := ctx.Config().GetInt("jpeg_quality", 75)
	overwrite := ctx.Config().GetBool("overwrite", false)

	var writer io.Writer
	if w := ctx.Writer(); w != nil {
		writer = w
	}

	return &imageOutlet{
		ctx:          ctx,
		dstPath:      path,
		dstPathField: dstPathField,
		srcFields:    fields,
		dstOverwrite: overwrite,
		dstWriter:    writer,
		jpegQuality:  jpegQuality,
	}
}

type imageOutlet struct {
	ctx          *engine.Context
	dstPathField string
	dstPath      string
	srcFields    []string
	dstSequence  int32
	dstOverwrite bool
	dstWriter    io.Writer
	// jpeg options
	jpegQuality int
}

func (iout *imageOutlet) Open() error {
	if iout.dstPath == "" && iout.dstPathField == "" {
		return fmt.Errorf("path and path_field are not specified")
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
		if field == nil || field.Type() != engine.BINARY {
			continue
		}
		srcContentType := engine.GetTagString(field.Tags, engine.CanonicalTagKey("Content-Type"))
		if !strings.HasPrefix(srcContentType, "image/") {
			continue
		}
		dstPath := iout.dstPath
		if iout.dstPathField != "" {
			pathField := rec.Field(iout.dstPathField)
			if pathField != nil {
				if path, ok := pathField.Value.String(); ok {
					dstPath = path
				}
			}
		}
		if err := iout.writeImageField(field, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func (iout *imageOutlet) writeImageField(field *engine.Field, dstPath string) error {
	if field.Type() != engine.BINARY {
		return fmt.Errorf("field %q is not binary", field.Name)
	}
	srcContentType := engine.GetTagString(field.Tags, engine.CanonicalTagKey("Content-Type"))
	if srcContentType == "" {
		return fmt.Errorf("field %q Content-Type is not specified", field.Name)
	}
	srcContentType = strings.Split(srcContentType, ";")[0]
	srcContentType = strings.TrimSpace(srcContentType)
	srcContentType = strings.ToLower(srcContentType)
	bv, _ := field.Value.Bytes()

	dstPattern := dstPath
	dstContentType := ""
	if dstPath != "" {
		ext := filepath.Ext(dstPath)
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
		ext = filepath.Ext(dstPath)
		base := filepath.Base(dstPath)
		baseWithoutExt := base[:len(base)-len(ext)]
		if iout.dstOverwrite {
			dstPattern = filepath.Join(filepath.Dir(dstPath), baseWithoutExt+ext)
		} else {
			dstPattern = filepath.Join(filepath.Dir(dstPath), baseWithoutExt+"_%d"+ext)
		}
	}
	iout.ctx.SetContentType(dstContentType)

	var srcImg image.Image
	switch srcContentType {
	default:
		return fmt.Errorf("field %q has unsupported Content-Type %q", field.Name, srcContentType)
	case "image/png":
		if dstContentType != "image/png" {
			if img, err := png.Decode(bytes.NewReader(bv)); err != nil {
				return err
			} else {
				srcImg = img
			}
		}
	case "image/jpeg":
		if dstContentType != "image/jpeg" {
			if img, err := jpeg.Decode(bytes.NewReader(bv)); err != nil {
				return err
			} else {
				srcImg = img
			}
		}
	case "image/gif":
		if dstContentType != "image/gif" {
			if img, err := gif.Decode(bytes.NewReader(bv)); err != nil {
				return err
			} else {
				srcImg = img
			}
		}
	case "image/bmp":
		if dstContentType != "image/bmp" {
			if img, err := bmp.Decode(bytes.NewReader(bv)); err != nil {
				return err
			} else {
				srcImg = img
			}
		}
	case "image/vnd.rgba":
		strStride := engine.GetTagString(field.Tags, engine.CanonicalTagKey("X-RGBA-Stride"))
		strRect := engine.GetTagString(field.Tags, engine.CanonicalTagKey("X-RGBA-Rectangle"))
		stride, err := strconv.ParseInt(strStride, 10, 64)
		if err != nil {
			return fmt.Errorf("field %q has invalid X-RGBA-Stride %q", field.Name, strStride)
		}
		rect, err := ParseImageRectangle(strRect)
		if err != nil {
			return fmt.Errorf("field %q has invalid X-RGBA-Rectangle, %s", field.Name, err.Error())
		}
		srcImg = &image.RGBA{Pix: bv, Stride: int(stride), Rect: rect}
	}

	var writer io.Writer
	if iout.dstWriter != nil {
		writer = iout.dstWriter
		goto write_image
	}
	if iout.dstOverwrite {
		dstPath := dstPattern
		if w, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
			return err
		} else {
			writer = w
			defer w.Close()
		}
		iout.ctx.LogDebug("outlets.image", "write_to", dstPath)
	} else {
		dstSequence := atomic.LoadInt32(&iout.dstSequence)
	inc_seq:
		if dstSequence > 0 {
			dstPath = fmt.Sprintf(dstPattern, dstSequence)
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
	var data []byte
	if srcImg == nil {
		// no need to convert
		data = bv
	} else {
		buff := &bytes.Buffer{}
		switch dstContentType {
		case "image/png":
			png.Encode(buff, srcImg)
		case "image/jpeg":
			jpeg.Encode(buff, srcImg, &jpeg.Options{Quality: iout.jpegQuality})
		case "image/gif":
			gif.Encode(buff, srcImg, nil)
		case "image/bmp":
			bmp.Encode(buff, srcImg)
		default:
			iout.ctx.LogError(fmt.Sprintf("unsupported image format [%s]", dstContentType))
		}
		data = buff.Bytes()
	}
	iout.ctx.SetContentLength(len(data))
	writer.Write(data)

	return nil
}
