//
// Copyright 2022 The Ent Authors.
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
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/go-openapi/runtime"
	"github.com/google/ent/utils"
	"github.com/in-toto/in-toto-golang/in_toto"
	"github.com/sigstore/cosign/cmd/cosign/cli/fulcio"
	"github.com/sigstore/cosign/pkg/cosign"
	"github.com/sigstore/cosign/pkg/signature"
	rekor "github.com/sigstore/rekor/pkg/client"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/client/index"
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/sigstore/rekor/pkg/types"
	hashedrekord "github.com/sigstore/rekor/pkg/types/hashedrekord/v0.0.1"
	intoto "github.com/sigstore/rekor/pkg/types/intoto/v0.0.1"
	rekord "github.com/sigstore/rekor/pkg/types/rekord/v0.0.1"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/signature/options"
	"github.com/spf13/cobra"
)

const (
	defaultRekorAddr = "https://rekor.sigstore.dev"

	GitHubActionsIssuer = "https://token.actions.githubusercontent.com"
	GitHubAccountIssuer = "https://github.com/login/oauth"
	GoogleAccountIssuer = "https://accounts.google.com"
)

var statusCmd = &cobra.Command{
	Use:  "status [digest]",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		digest, err := utils.ParseDigest(args[0])
		if err != nil {
			log.Fatalf("could not parse digest: %v", err)
			return
		}
		status(digest)
	},
}

// rekorStatus checks if the digest is signed by one or more rekor entries, and prints out their
// details. It only handles entries that root into Fulcio.
//
// To sign an entry locally:
//
// COSIGN_EXPERIMENTAL=1 cosign sign-blob ./README.md
func rekorStatus(digest utils.Digest) {
	rc, err := rekor.GetRekorClient(defaultRekorAddr, rekor.WithUserAgent("ent"))
	if err != nil {
		log.Fatalf("could not create rekor client: %v", err)
		return
	}
	params := index.NewSearchIndexParams()
	params.Query = &models.SearchIndex{Hash: string(digest)}
	res, err := rc.Index.SearchIndex(params)
	if err != nil {
		log.Fatalf("could not search rekor index: %v", err)
		return
	}
	log.Printf("found %d rekor entries", len(res.Payload))
	for _, id := range res.Payload {
		p := entries.NewGetLogEntryByUUIDParams()
		p.EntryUUID = id
		e, err := rc.Entries.GetLogEntryByUUID(p)
		if err != nil {
			log.Fatalf("could not get rekor entry: %v", err)
			return
		}
		en := e.Payload[id]
		_, err = certs(&en)
		if err != nil {
			log.Fatalf("could not get certs from rekor entry %q: %v", id, err)
			return
		}
	}
}

func certs(e *models.LogEntryAnon) ([]*x509.Certificate, error) {
	log.Printf("log entry %q", *e.LogID)
	b, err := base64.StdEncoding.DecodeString(e.Body.(string))
	if err != nil {
		return nil, err
	}
	log.Printf("log entry %s", string(b))
	pe, err := models.UnmarshalProposedEntry(bytes.NewReader(b), runtime.JSONConsumer())
	if err != nil {
		return nil, err
	}
	eimpl, err := types.NewEntry(pe)
	if err != nil {
		return nil, err
	}
	// log.Printf("entry: %+v", e)
	// log.Printf("entry: %+v", eimpl)
	var (
		publicKey []byte
		data      []byte
		sig       []byte
	)
	switch ei := eimpl.(type) {
	case *rekord.V001Entry:
		publicKey = ei.RekordObj.Signature.PublicKey.Content
		data = ei.RekordObj.Data.Content
		sig = ei.RekordObj.Signature.Content
	case *hashedrekord.V001Entry:
		publicKey = ei.HashedRekordObj.Signature.PublicKey.Content
		data, err = hex.DecodeString(*ei.HashedRekordObj.Data.Hash.Value)
		if err != nil {
			return nil, err
		}
		sig = ei.HashedRekordObj.Signature.Content
	case *intoto.V001Entry:
		var i in_toto.ProvenanceStatement
		attestationBytes, err := base64.StdEncoding.DecodeString(string(e.Attestation.Data))
		if err != nil {
			return nil, err
		}
		err = json.NewDecoder(bytes.NewReader(attestationBytes)).Decode(&i)
		if err != nil {
			return nil, err
		}
		log.Printf("intoto attestation %+v", i)
		publicKey = *ei.IntotoObj.PublicKey
		// TODO
	default:
		return nil, errors.New("unexpected tlog entry type")
	}
	// log.Printf("data: (%d) %v", len(data), data)
	// log.Printf("sig: (%d) %v", len(sig), sig)
	// log.Printf("public key: (%d) %v", len(publicKey), string(publicKey))

	certs, err := cryptoutils.UnmarshalCertificatesFromPEM(publicKey)
	if err != nil {
		block, _ := pem.Decode(publicKey)
		pk, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("could not parse public key: %v", err)
		}
		log.Printf("public key: %+v", pk)
		switch pk := pk.(type) {
		case *ecdsa.PublicKey:
			ok := ecdsa.VerifyASN1(pk, data, sig)
			log.Printf("ecdsa verification: %v", ok)
		default:
			return nil, fmt.Errorf("unexpected public key type: %T", pk)
		}
		return nil, nil
	} else {

		if len(certs) == 0 {
			return nil, errors.New("no certs found in pem tlog")
		}

		co := &cosign.CheckOpts{
			RootCerts: fulcio.GetRoots(),
			// CertOidcIssuer: "https://token.actions.githubusercontent.com",
			// CertOidcIssuer: "https://github.com/login/oauth",
			// CertOidcIssuer: "https://accounts.google.com",
		}
		for _, c := range certs {
			verifier, err := cosign.ValidateAndUnpackCert(c, co)
			if err != nil {
				log.Printf("could not validate cert: %v", err)
				continue
			}
			err = verifier.VerifySignature(bytes.NewReader(sig), nil, options.WithDigest(data))
			if err != nil {
				log.Printf("could not verify signature: %v", err)
				continue
			}
			log.Printf("verified signature")
			log.Printf("cert OIDC issuer: %q", signature.CertIssuerExtension(c))
			log.Printf("cert OIDC subject: %q", signature.CertSubject(c))
		}
	}

	return certs, err
}

func status(digest utils.Digest) {
	config := readConfig()
	s := []<-chan string{}
	for _, remote := range config.Remotes {
		c := make(chan string)
		s = append(s, c)
		go func(remote Remote, c chan<- string) {
			objectGetter := getObjectGetter(remote)
			marker := color.GreenString("✓")
			_, err := objectGetter.Get(context.Background(), digest)
			if err != nil {
				marker = color.RedString("✗")
			}
			c <- fmt.Sprintf("%s %s [%s]\n", color.YellowString(string(digest)), marker, remote.Name)
		}(remote, c)
	}
	for _, c := range s {
		fmt.Printf(<-c)
	}
	rekorStatus(digest)
}
