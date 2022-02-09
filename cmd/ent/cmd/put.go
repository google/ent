package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var putCmd = &cobra.Command{
	Use:  "put [filename]",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := ""
		if len(args) > 0 {
			filename = args[0]
		}
		if filename == "" {
			err := putStdin()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			err := putFile(filename)
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

func putStdin() error {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("could not read stdin: %v", err)
	}
	return putData(data)
}

func putFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read file %q: %v", filename, err)
	}
	return putData(data)
}

func putData(data []byte) error {
	localHash := utils.ComputeHash(data)
	if exists(localHash) {
		marker := color.GreenString("✓")
		fmt.Printf("%s %s\n", color.YellowString(string(localHash)), marker)
		return nil
	} else {
		config := readConfig()
		nodeService := getObjectStore(config)
		_, err := nodeService.Put(context.Background(), data)
		if err != nil {
			return fmt.Errorf("could not add object: %v", err)
		}
		marker := color.BlueString("↑")
		fmt.Printf("%s %s\n", color.YellowString(string(localHash)), marker)
		return nil
	}
}

func exists(hash utils.Hash) bool {
	config := readConfig()
	nodeService := getObjectStore(config)
	_, err := nodeService.Get(context.Background(), hash)
	if err == nodeservice.ErrNotFound {
		return false
	} else if err != nil {
		log.Fatal(err)
	}
	return true
}
