package remote

import (
	"fmt"
	"log"
	"net/url"

	"github.com/google/ent/cmd/ent/config"
	"github.com/google/ent/nodeservice"
	pb "github.com/google/ent/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
		parsedURL, err := url.Parse(remote.URL)
		if err != nil {
			log.Fatalf("failed to parse url: %v", err)
		}

		o := []grpc.DialOption{}
		if parsedURL.Scheme == "http" {
			o = append(o, grpc.WithInsecure())
		} else {
			o = append(o, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
		}
		port := parsedURL.Port()
		if port == "" {
			if parsedURL.Scheme == "http" {
				port = "80"
			} else {
				port = "443"
			}
		}
		cc, err := grpc.Dial(parsedURL.Hostname()+":"+port, o...)
		if err != nil {
			log.Fatalf("failed to dial: %v", err)
		}
		client := pb.NewEntClient(cc)
		return &nodeservice.Remote{
			APIURL: remote.URL,
			APIKey: remote.APIKey,
			GRPC:   client,
		}
	} else {
		return nil
	}
}
