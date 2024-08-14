package util_test

import (
	"fmt"

	"github.com/OutOfBedlam/tine/util"
)

func ExampleCamelToSnake() {
	fmt.Println(util.CamelToSnake("CamelToSnake"))
	// Output:
	// camel_to_snake
}

func ExampleCamelToKebab() {
	fmt.Println(util.CamelToKebab("CamelToSnake"))
	// Output:
	// camel-to-snake
}
