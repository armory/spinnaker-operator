// +build no-test

package main

import (
	"fmt"
	"github.com/operator-framework/operator-sdk/cmd/operator-sdk/add"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	root := &cobra.Command{
		Use: "operator-sdk",
	}
	if len(os.Args) == 1 {
		fmt.Println("Usage: go run tools/add.go [NEW_API_VERSION]")
		os.Exit(1)
	}
	root.AddCommand(add.NewCmd())
	root.SetArgs([]string{"add", "api", fmt.Sprintf("--api-version=spinnaker.io/%s", os.Args[1]),
		"--kind=SpinnakerService"})
	if err := root.Execute(); err != nil {
		os.Exit(2)
	}
}
