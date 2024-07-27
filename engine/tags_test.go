package engine_test

import (
	"fmt"
	"slices"

	"github.com/OutOfBedlam/tine/engine"
)

func ExampleTags_Get() {
	tags := engine.Tags{
		"key1": engine.NewValue("value1"),
		"key2": engine.NewValue("value2"),
	}
	str, _ := tags.Get("key1").String()
	fmt.Println(str)
	fmt.Println(tags.Get("key3"))
	// Output:
	// value1
	// <nil>
}

func ExampleTags_Set() {
	tags := engine.Tags{
		"key1": engine.NewValue("value1"),
		"key2": engine.NewValue("value2"),
	}
	tags.Set("key1", engine.NewValue("value3"))
	str, _ := tags.Get("key1").String()
	fmt.Println(str)
	// Output:
	// value3
}

func ExampleTags_Del() {
	tags := engine.Tags{
		"key1": engine.NewValue("value1"),
		"key2": engine.NewValue("value2"),
	}
	tags.Del("key1")
	fmt.Println(tags.Get("key1"))
	// Output:
	// <nil>
}

func ExampleTags_Clone() {
	tags := engine.Tags{
		"key1": engine.NewValue("value1"),
		"key2": engine.NewValue("value2"),
	}
	tags2 := tags.Clone()
	str, _ := tags2.Get("key1").String()
	fmt.Println(str)
	// Output:
	// value1
}

func ExampleTags_Clone_empty() {
	tags := engine.Tags{}
	tags2 := tags.Clone()
	fmt.Println(tags2)
	// Output:
	// map[]
}

func ExampleTags_Clear() {
	tags := engine.Tags{
		"key1": engine.NewValue("value1"),
		"key2": engine.NewValue("value2"),
	}
	tags.Clear()
	fmt.Println(tags)
	// Output:
	// map[]
}

func ExampleTags_Len() {
	tags := engine.Tags{
		"key1": engine.NewValue("value1"),
		"key2": engine.NewValue("value2"),
	}
	fmt.Println(tags.Len())
	// Output:
	// 2
}

func ExampleTags_IsEmpty() {
	tags := engine.Tags{}
	fmt.Println(tags.IsEmpty())
	// Output:
	// true
}

func ExampleTags_IsNotEmpty() {
	tags := engine.Tags{
		"key1": engine.NewValue("value1"),
		"key2": engine.NewValue("value2"),
	}
	fmt.Println(tags.IsNotEmpty())
	// Output:
	// true
}

func ExampleTags_Merge() {
	tags := engine.Tags{
		"key1": engine.NewValue("value1"),
		"key2": engine.NewValue("value2"),
	}
	tags2 := engine.Tags{
		"key3": engine.NewValue("value3"),
		"key4": engine.NewValue("value4"),
	}
	tags.Merge(tags2)
	str, _ := tags.Get("key3").String()
	fmt.Println(str)
	// Output:
	// value3
}

func ExampleTags_Keys() {
	tags := engine.Tags{
		"key1": engine.NewValue("value1"),
		"key2": engine.NewValue("value2"),
	}
	keys := tags.Keys()
	slices.Sort(keys)
	fmt.Println(keys)
	// Output:
	// [key1 key2]
}

func ExampleTags_Keys_empty() {
	tags := engine.Tags{}
	keys := tags.Keys()
	fmt.Println(keys)
	// Output:
	// []
}

func ExampleTags_MergeWithPrefix() {
	tags := engine.Tags{
		"key1": engine.NewValue("value1"),
		"key2": engine.NewValue("value2"),
	}
	tags2 := engine.Tags{
		"key3": engine.NewValue("value3"),
		"key4": engine.NewValue("value4"),
	}
	tags.MergeWithPrefix(tags2, "prefix.")
	str, _ := tags.Get("prefix.key3").String()
	fmt.Println(str)
	// Output:
	// value3
}

func ExampleTags_MergeWithSuffix() {
	tags := engine.Tags{
		"key1": engine.NewValue("value1"),
		"key2": engine.NewValue("value2"),
	}
	tags2 := engine.Tags{
		"key3": engine.NewValue("value3"),
		"key4": engine.NewValue("value4"),
	}
	tags.MergeWithSuffix(tags2, ".suffix")
	str, _ := tags.Get("key3.suffix").String()
	fmt.Println(str)
	// Output:
	// value3
}

func ExampleTags_MergeWithPrefixSuffix() {
	tags := engine.Tags{
		"key1": engine.NewValue("value1"),
		"key2": engine.NewValue("value2"),
	}
	tags2 := engine.Tags{
		"key3": engine.NewValue("value3"),
		"key4": engine.NewValue("value4"),
	}
	tags.MergeWithPrefixSuffix(tags2, "prefix.", ".suffix")
	str, _ := tags.Get("prefix.key3.suffix").String()
	fmt.Println(str)
	// Output:
	// value3
}
