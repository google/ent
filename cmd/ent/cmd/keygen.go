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

package cmd

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"os"

	"github.com/google/ent/log"
	"github.com/spf13/cobra"
)

var keygenCmd = &cobra.Command{
	Use: "keygen",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			log.Criticalf(ctx, "generate key pair: %v", err)
			os.Exit(1)
		}
		sk, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			log.Criticalf(ctx, "marshal private key: %v", err)
			os.Exit(1)
		}
		sks := base64.URLEncoding.EncodeToString(sk)
		log.Infof(ctx, "Secret key: %q", sks)

		pk, err := x509.MarshalPKIXPublicKey(&k.PublicKey)
		if err != nil {
			log.Criticalf(ctx, "marshal public key: %v", err)
			os.Exit(1)
		}
		pks := base64.URLEncoding.EncodeToString(pk)
		log.Infof(ctx, "Public key: %q", pks)

		text := "hello world"
		sig, err := ecdsa.SignASN1(rand.Reader, k, []byte(text))
		if err != nil {
			log.Criticalf(ctx, "sign ASN1: %v", err)
			os.Exit(1)
		}
		sigs := base64.URLEncoding.EncodeToString(sig)
		log.Infof(ctx, "Signature: %q", sigs)

		// verify
		ok := ecdsa.VerifyASN1(&k.PublicKey, []byte(text), sig)
		log.Infof(ctx, "Verify: %v", ok)
	},
}

func init() {
	keygenCmd.PersistentFlags().StringVar(&remoteFlag, "remote", "", "remote")
}
