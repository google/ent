package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:  "tags",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tags, err := tagStore.List(context.Background())
		if err != nil {
			log.Fatalf("could not list tags: %v", err)
		}
		for _, tag := range tags {
			tagValue, err := tagStore.Get(context.Background(), tag)
			if err != nil {
				log.Fatalf("could not list tags: %v", err)
			}
			fmt.Printf("%s %s\n", color.YellowString(string(tagValue)), tag)
		}
	},
}
