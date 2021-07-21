package cmd

import (
	"log"
	"path/filepath"

	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
)

var makeCmd = &cobra.Command{
	Use:  "make [target directory]",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetDir := "."
		if len(args) >= 1 {
			targetDir = args[0]
		}
		targetDir, err := filepath.Abs(targetDir)
		if err != nil {
			log.Fatalf("could not normalize target directory %q, %v", targetDir, err)
		}

		plan, err := parsePlan(filepath.Join(targetDir, planFilename))
		if err != nil {
			log.Fatalf("could not parse plan: %v", err)
		}
		log.Printf("plan: %#v", plan)

		for _, o := range plan.Overrides {
			base, err := cid.Decode(o.From)
			if err != nil {
				log.Fatalf("could not decode cid: %v", err)
			}
			pull(base, o.Path, o.Executable)
		}
	},
}
