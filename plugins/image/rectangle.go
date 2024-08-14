package image

import (
	"errors"
	"image"
	"strconv"
	"strings"
)

func ParseImageRectangle(strRect string) (image.Rectangle, error) {
	var err error
	rect := image.Rect(0, 0, 0, 0)
	points := strings.Split(strRect, ",")
	if len(points) != 4 {
		return rect, errors.New("invalid rectangle format")
	}
	rect.Min.X, err = strconv.Atoi(points[0])
	if err != nil {
		return rect, errors.New("invalid min x")
	}
	rect.Min.Y, err = strconv.Atoi(points[1])
	if err != nil {
		return rect, errors.New("invalid min y")
	}
	rect.Max.X, err = strconv.Atoi(points[2])
	if err != nil {
		return rect, errors.New("invalid max x")
	}
	rect.Max.Y, err = strconv.Atoi(points[3])
	if err != nil {
		return rect, errors.New("invalid max y")
	}
	return rect, nil
}
