package util

import "fmt"

var byteSizes = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

func FormatFileSizeInt(bytes int) string {
	return FormatFileSize(int64(bytes))
}

// ByteSize returns a human-readable byte size string.
func FormatFileSize(bytes int64) string {
	if bytes == 0 {
		return "0B"
	}
	s := float64(bytes)
	i := 0
	for s >= 1024 && i < len(byteSizes)-1 {
		s /= 1024
		i++
	}

	f := "%.0f %s"
	if i > 1 {
		f = "%.2f %s"
	}
	return fmt.Sprintf(f, s, byteSizes[i])
}
