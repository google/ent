package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/google/ent/datastore"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/objectstore"
	"github.com/google/ent/tagstore"
	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/cobra"
)

var (
	nodeService nodeservice.NodeService
	tagStore    tagstore.TagStore
)

type Config struct {
	DefaultRemote string `toml:"default_remote"`
	Remotes       map[string]Remote
}

type Remote struct {
	Path string
	URL  string
}

type Plan struct {
	Overrides []Override
}

type Override struct {
	Path       string
	From       string
	Executable bool
}

const planFilename = "entplan.toml"

func parsePlan(filename string) (Plan, error) {
	var plan Plan

	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return plan, err
	}
	err = toml.Unmarshal(f, &plan)
	if err != nil {
		return plan, err
	}

	return plan, nil
}

func parseIgnore(targetDir string) *ignore.GitIgnore {
	s, err := os.Stat(targetDir)
	if err != nil {
		log.Panic(err)
	}
	if !s.IsDir() {
		return &ignore.GitIgnore{}
	}
	filename := filepath.Join(targetDir, ".gitignore")
	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		return &ignore.GitIgnore{}
	} else if err != nil {
		log.Panic(err)
	}

	i, err := ignore.CompileIgnoreFile(filename)
	if err != nil {
		log.Panic(err)
	}
	return i
}

func InitRemote(remote Remote) {
	if remote.URL != "" {
		nodeService = nodeservice.Remote{
			APIURL: remote.URL,
		}
	} else if remote.Path != "" {
		baseDir := remote.Path

		{
			blobsDir := filepath.Join(baseDir, "blobs")
			err := os.MkdirAll(blobsDir, 0755)
			if err != nil {
				log.Fatalf("could not create blobs dir: %v", err)
			}
			nodeService = nodeservice.DataStore{
				Inner: objectstore.Store{
					Inner: datastore.File{
						DirName: blobsDir,
					},
				},
			}
		}

		{
			tagsDir := filepath.Join(baseDir, "tags")
			err := os.MkdirAll(tagsDir, 0755)
			if err != nil {
				log.Fatalf("could not create tags dir: %v", err)
			}
			tagStore = tagstore.File{
				DirName: tagsDir,
			}
		}
	} else {
		log.Fatal("no remote specified")
	}
}

var rootCmd = &cobra.Command{
	Use: "ent",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		s, err := os.UserConfigDir()
		if err != nil {
			log.Fatalf("could not load config dir: %v", err)
		}
		s = filepath.Join(s, "ent.toml")
		f, err := ioutil.ReadFile(s)
		if err != nil {
			log.Printf("could not read config: %v", err)
			// Continue anyways.
		}
		config := Config{}
		err = toml.Unmarshal(f, &config)
		if err != nil {
			log.Fatalf("could not parse config: %v", err)
		}
		// log.Printf("parsed config: %#v", config)

		if remoteName == "" && config.DefaultRemote != "" {
			remoteName = config.DefaultRemote
		}

		remote, ok := config.Remotes[remoteName]
		if !ok {
			log.Fatalf("Invalid remote name: %q", remoteName)
		}
		InitRemote(remote)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var (
	remoteName string
	tagName    string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&remoteName, "remote", "", "")

	pushCmd.Flags().StringVar(&tagName, "tag", "", "")

	rootCmd.AddCommand(catCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(makeCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(tagsCmd)
}

func traverse(base string, relativeFilename string, i *ignore.GitIgnore, f func(string, format.Node) error) cid.Cid {
	file, err := os.Open(path.Join(base, relativeFilename))
	if err != nil {
		log.Fatal(err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if fileInfo.IsDir() {
		files, err := file.Readdir(-1)
		if err != nil {
			log.Fatal(err)
		}

		node := utils.NewProtoNode()
		for _, ff := range files {
			newRelativeFilename := path.Join(relativeFilename, ff.Name())
			if i.MatchesPath(newRelativeFilename) {
				// Nothing
			} else {
				hash := traverse(base, newRelativeFilename, i, f)
				utils.SetLink(node, ff.Name(), hash)
			}
		}

		err = f(relativeFilename, node)
		if err != nil {
			log.Fatal(err)
		}

		return node.Cid()
		// } else if fileInfo.Mode() == os.ModeSymlink {
		// skip
	} else {
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}
		node, err := utils.ParseRawNode(bytes)
		if err != nil {
			log.Fatal(err)
		}

		err = f(relativeFilename, node)
		if err != nil {
			log.Fatal(err)
		}

		return node.Cid()
	}
}
