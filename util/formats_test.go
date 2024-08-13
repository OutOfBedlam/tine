package util_test

import (
	"fmt"

	"github.com/OutOfBedlam/tine/util"
)

func ExampleFormatFileSize() {
	fmt.Println(util.FormatFileSizeInt(0))
	fmt.Println(util.FormatFileSize(1))
	fmt.Println(util.FormatFileSizeInt(1024))
	fmt.Println(util.FormatFileSize(1024 * 1024))
	fmt.Println(util.FormatFileSize(1024 * 1024 * 1024))
	fmt.Println(util.FormatFileSize(1024 * 1024 * 1024 * 1024))
	fmt.Println(util.FormatFileSize(1024 * 1024 * 1024 * 1024 * 1024))
	// Output:
	// 0B
	// 1B
	// 1.00KB
	// 1.00MB
	// 1.00GB
	// 1.00TB
	// 1.00PB
}

func ExampleFormatCount() {
	fmt.Println(util.FormatCount(0, util.CountUnitLines))
	fmt.Println(util.FormatCount(1, util.CountUnitLines))
	fmt.Println(util.FormatCount(2, util.CountUnitLines))
	// Output:
	// 0 lines
	// 1 line
	// 2 lines
}
