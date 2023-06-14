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
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"

	"github.com/golang/protobuf/proto"
	pb "github.com/google/ent/proto"
	"github.com/spf13/cobra"
)

var MapCmd = &cobra.Command{
	Use: "map",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	MapCmd.AddCommand(setCmd)
	MapCmd.AddCommand(getCmd)
}

func ValidateRequest(m *pb.MapSetRequest) error {
	pk, err := x509.ParsePKIXPublicKey(m.PublicKey)
	if err != nil {
		return err
	}
	ecpk, ok := pk.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("invalid public key")
	}

	entryBytes, err := proto.Marshal(m.Entry)
	if err != nil {
		return err
	}
	if !ecdsa.VerifyASN1(ecpk, entryBytes, m.EntrySignature) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}
