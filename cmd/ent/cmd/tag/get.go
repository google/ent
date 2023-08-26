//
// Copyright 2023 The Ent Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tag

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/google/ent/cmd/ent/config"
	"github.com/google/ent/cmd/ent/remote"
	pb "github.com/google/ent/proto"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use: "get",
	Run: func(cmd *cobra.Command, args []string) {
		c := config.ReadConfig()

		pkb, err := base64.URLEncoding.DecodeString(publicKey)
		if err != nil {
			log.Fatalf("failed to decode public key: %v", err)
		}
		pk, err := x509.ParsePKIXPublicKey(pkb)
		if err != nil {
			log.Fatalf("failed to parse public key: %v", err)
		}
		ecpk, ok := pk.(*ecdsa.PublicKey)
		if !ok {
			log.Fatalf("public key is not ECDSA")
		}
		log.Printf("public key: %v", ecpk)

		req := pb.GetTagRequest{
			PublicKey: pkb,
			Tag:       tag,
		}
		log.Printf("request: %+v", &req)

		r := c.Remotes[0]
		nodeService := remote.GetObjectStore(r)
		ctx := context.Background()
		res, err := nodeService.GRPC.GetTag(ctx, &req)
		if err != nil {
			log.Fatalf("failed to get: %v", err)
		}
		log.Printf("response: %+v", res)

		digest := utils.DigestFromProto(res.Entry.Target)
		out := utils.DigestToHumanString(digest)
		fmt.Printf("%s\n", out)
	},
}

func init() {
	getCmd.PersistentFlags().StringVar(&publicKey, "public-key", "", "public key")
	getCmd.PersistentFlags().StringVar(&tag, "tag", "", "tag")
}
