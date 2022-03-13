package main

import (
	"os"

	"github.com/coder/flog"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:     "github-enricher",
		Short:   "Enrich GitHub data",
		Example: "github-enricher < input.csv > output.csv",
		Run: func(cmd *cobra.Command, args []string) {
			e := enricherEngine{
				Log: flog.New(),
			}
			err := e.Run(os.Stdout, os.Stdin)
			if err != nil {
				flog.Fatal("%v+", err)
			}
		},
	}
	err := cmd.Execute()
	if err != nil {
		flog.Fatal(err.Error())
	}
}
