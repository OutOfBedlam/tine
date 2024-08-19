package util_test

import (
	"fmt"

	"github.com/OutOfBedlam/tine/util"
)

func ExampleJsonFloat_MarshalJSON() {
	jf := util.JsonFloat{Value: 3.1415926535, Decimal: 2}
	b, err := jf.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	// Output:
	// 3.14
}
