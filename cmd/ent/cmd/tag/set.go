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
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/google/ent/cmd/ent/config"
	"github.com/google/ent/cmd/ent/remote"
	"github.com/google/ent/log"
	pb "github.com/google/ent/proto"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var (
	publicKey string
	label     string
	target    string
)

var setCmd = &cobra.Command{
	Use:  "set",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		c := config.ReadConfig()

		skb, err := base64.URLEncoding.DecodeString(c.SecretKey)
		if err != nil {
			log.Criticalf(ctx, "decode secret key: %v", err)
			os.Exit(1)
		}
		sk, err := x509.ParseECPrivateKey(skb)
		if err != nil {
			log.Criticalf(ctx, "parse secret key: %v", err)
			os.Exit(1)
		}

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
		// Compare public key with secret key
		if !ecpk.Equal(sk.Public()) {
			log.Criticalf(ctx, "public key mismatch")
			os.Exit(1)
		}

		targetDigest, err := utils.ParseDigest(target)
		if err != nil {
			log.Criticalf(ctx, "parse target: %v", err)
			os.Exit(1)
		}
		targetProto := utils.DigestToProto(targetDigest)

		tag := pb.Tag{
			Label:  label,
			Target: targetProto,
		}
		tagBytes, err := proto.Marshal(&tag)
		if err != nil {
			log.Criticalf(ctx, "marshal tag: %v", err)
			os.Exit(1)
		}
		signature, err := ecdsa.SignASN1(rand.Reader, sk, tagBytes)
		if err != nil {
			log.Criticalf(ctx, "sign tag: %v", err)
			os.Exit(1)
		}
		log.Infof(ctx, "signature: %s", base64.URLEncoding.EncodeToString(signature))

		req := pb.SetTagRequest{
			SignedTag: &pb.SignedTag{
				Tag:          &tag,
				TagSignature: signature,
				PublicKey:    pkb,
			},
		}
		log.Infof(ctx, "request: %+v", &req)

		err = ValidateEntry(req.SignedTag.Tag, ecpk, req.SignedTag.TagSignature)
		if err != nil {
			log.Criticalf(ctx, "validate tag: %v", err)
			os.Exit(1)
		}

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
		_, err = nodeService.GRPC.SetTag(ctx, &req)
		if err != nil {
			log.Criticalf(ctx, "set tag: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	setCmd.PersistentFlags().StringVar(&publicKey, "public-key", "", "public key")
	setCmd.PersistentFlags().StringVar(&label, "label", "", "label")
	setCmd.PersistentFlags().StringVar(&target, "target", "", "target")
	setCmd.PersistentFlags().StringVar(&remoteFlag, "remote", "", "remote")
}
