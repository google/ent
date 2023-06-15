package remote

import (
	"fmt"

	"github.com/google/ent/cmd/ent/config"
	"github.com/google/ent/nodeservice"
)

func GetRemote(c config.Config, remoteName string) (config.Remote, error) {
	for _, remote := range c.Remotes {
		if remote.Name == remoteName {
			return remote, nil
		}
	}
	return config.Remote{}, fmt.Errorf("remote %q not found", remoteName)
}

func GetObjectStore(remote config.Remote) *nodeservice.Remote {
	if remote.Write {
		return &nodeservice.Remote{
			APIURL: remote.URL,
			APIKey: remote.APIKey,
		}
	} else {
		return nil
	}
}
