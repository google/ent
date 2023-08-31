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
	"os"

	"github.com/google/ent/cmd/ent/config"
	"github.com/google/ent/cmd/ent/remote"
	"github.com/google/ent/log"
	pb "github.com/google/ent/proto"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use: "get",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		c := config.ReadConfig()

		pkb, err := base64.URLEncoding.DecodeString(publicKey)
		if err != nil {
			log.Criticalf(ctx, "decode public key: %v", err)
			os.Exit(1)
		}
		pk, err := x509.ParsePKIXPublicKey(pkb)
		if err != nil {
			log.Criticalf(ctx, "parse public key: %v", err)
			os.Exit(1)
		}
		ecpk, ok := pk.(*ecdsa.PublicKey)
		if !ok {
			log.Criticalf(ctx, "public key is not ECDSA")
			os.Exit(1)
		}
		log.Infof(ctx, "public key: %v", ecpk)

		req := pb.GetTagRequest{
			PublicKey: pkb,
			Label:     label,
		}
		log.Infof(ctx, "request: %+v", &req)

		r := c.Remotes[0]
		if remoteFlag != "" {
			var err error
			r, err = remote.GetRemote(c, remoteFlag)
			if err != nil {
				log.Criticalf(ctx, "could not use remote: %v", err)
				os.Exit(1)
			}
		}
		log.Debugf(ctx, "using remote %q", r.Name)

		nodeService := remote.GetObjectStore(r)
		res, err := nodeService.GRPC.GetTag(ctx, &req)
		if err != nil {
			log.Criticalf(ctx, "failed to get: %v", err)
			os.Exit(1)
		}
		log.Infof(ctx, "response: %+v", res)

		digest := utils.DigestFromProto(res.SignedTag.Tag.Target)
		out := utils.DigestToHumanString(digest)
		fmt.Printf("%s\n", out)
	},
}

func init() {
	getCmd.PersistentFlags().StringVar(&publicKey, "public-key", "", "public key")
	getCmd.PersistentFlags().StringVar(&label, "label", "", "label")
	getCmd.PersistentFlags().StringVar(&remoteFlag, "remote", "", "remote")
}
