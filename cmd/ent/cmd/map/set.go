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

package _map

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/google/ent/cmd/ent/config"
	"github.com/google/ent/cmd/ent/remote"
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
		c := config.ReadConfig()
		skb, err := base64.URLEncoding.DecodeString(c.SecretKey)
		if err != nil {
			log.Fatalf("failed to decode secret key: %v", err)
		}
		sk, err := x509.ParseECPrivateKey(skb)
		if err != nil {
			log.Fatalf("failed to parse secret key: %v", err)
		}

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
		// Compare public key with secret key
		if !ecpk.Equal(sk.Public()) {
			log.Fatalf("public key mismatch")
		}

		targetDigest, err := utils.ParseDigest(target)
		if err != nil {
			log.Fatalf("failed to parse target: %v", err)
		}
		targetProto := utils.DigestToProto(targetDigest)

		entry := pb.MapSetRequest_Entry{
			Label:  label,
			Target: targetProto,
		}
		entryBytes, err := proto.Marshal(&entry)
		if err != nil {
			log.Fatalf("failed to marshal entry: %v", err)
		}
		signature, err := ecdsa.SignASN1(rand.Reader, sk, entryBytes)
		if err != nil {
			log.Fatalf("failed to sign: %v", err)
		}
		log.Printf("signature: %s", base64.URLEncoding.EncodeToString(signature))

		req := pb.MapSetRequest{
			Entry:          &entry,
			PublicKey:      pkb,
			EntrySignature: signature,
		}
		log.Printf("request: %+v", &req)

		err = ValidateRequest(&req)
		if err != nil {
			log.Fatalf("failed to validate map: %v", err)
		}

		config := config.ReadConfig()
		r := config.Remotes[0]
		nodeService := remote.GetObjectStore(r)
		ctx := context.Background()
		err = nodeService.MapSet(ctx, &req)
		log.Printf("err: %v", err)
	},
}

func init() {
	setCmd.PersistentFlags().StringVar(&publicKey, "public-key", "", "public key")
	setCmd.PersistentFlags().StringVar(&label, "label", "", "label")
	setCmd.PersistentFlags().StringVar(&target, "target", "", "target")
}
