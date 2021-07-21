package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/google/ent/nodeservice"
	format "github.com/ipfs/go-ipld-format"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:  "status",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := "."
		if len(args) > 0 {
			target = args[0]
		}
		i := parseIgnore(target)
		traverse(target, "", i, status)
	},
}

func status(filename string, node format.Node) error {
	if filename == "" {
		filename = "."
	}
	marker := color.RedString("*")
	_, err := nodeService.GetObject(context.Background(), node.Cid().Hash())
	if err == nodeservice.ErrNotFound {
		marker = color.RedString("*")
	} else if err != nil {
		log.Fatalf("could not fetch object: %v", err)
	} else {
		marker = color.GreenString("✓")
	}
	// hash = hash[len(hash)-16:]
	c := node.Cid().String()
	fmt.Printf("%s %s %s\n", color.YellowString(c), marker, filename)
	// h := node.Cid().Hash()
	// fmt.Printf("%s %s %s\n", color.YellowString(h.HexString()), marker, filename)
	// fmt.Printf("%s %s %s\n", color.YellowString(h.B58String()), marker, filename)
	return nil
}
