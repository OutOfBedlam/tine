package main

import (
	"context"

	"github.com/OutOfBedlam/tine/cmd"
	"github.com/spf13/cobra"
)

func main() {
	cobra.CheckErr(cmd.NewCmd().ExecuteContext(context.Background()))
}
