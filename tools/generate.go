// +build no-test

package main

import (
	"fmt"
	"github.com/operator-framework/operator-sdk/cmd/operator-sdk/generate"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	root := &cobra.Command{
		Use: "operator-sdk",
	}
	if len(os.Args) == 1 {
		fmt.Println("Usage: go run tools/generate.go [k8s|openapi]")
		os.Exit(1)
	}
	root.AddCommand(generate.NewCmd())
	root.SetArgs([]string{"generate", os.Args[1]})
	if err := root.Execute(); err != nil {
		os.Exit(2)
	}
}
