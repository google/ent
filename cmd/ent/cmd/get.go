package cmd

import (
	"context"
	"log"
	"os"

	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

func get(hash utils.Hash) {
	config := readConfig()
	objectGetter := getObjectGetter(config)
	object, err := objectGetter.Get(context.Background(), hash)
	if err != nil {
		log.Fatalf("could not download target: %s", err)
	}
	os.Stdout.Write(object)
}

var getCmd = &cobra.Command{
	Use:  "get [hash] [path]",
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		hash, err := utils.ParseHash(args[0])
		if err != nil {
			log.Fatalf("could not parse hash: %v", err)
			return
		}
		get(hash)
	},
}
