//
// Copyright 2021 The Ent Authors.
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

package nodeservice

import (
	"context"
	"fmt"
	"io"

	"github.com/google/ent/log"
	pb "github.com/google/ent/proto"
	"github.com/google/ent/utils"
	"github.com/schollz/progressbar/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type Remote struct {
	APIURL string
	APIKey string
	GRPC   pb.EntClient
}

const (
	APIKeyHeader = "x-api-key"
	chunkSize    = 1024 * 1024
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

func (s Remote) Get(ctx context.Context, digest utils.Digest) ([]byte, error) {
	md := metadata.New(nil)
	md.Set(APIKeyHeader, s.APIKey)
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := pb.GetEntryRequest{
		Digest: utils.DigestToProto(digest),
	}
	c, err := s.GRPC.GetEntry(ctx, &req)
	if err != nil {
		return nil, err
	}

	res, err := c.Recv()
	if err != nil {
		return nil, err
	}

	return res.GetChunk().Data, nil
}

func (s Remote) Put(ctx context.Context, size uint64, r io.Reader) (utils.Digest, error) {
	md := metadata.New(nil)
	md.Set(APIKeyHeader, s.APIKey)
	ctx = metadata.NewOutgoingContext(ctx, md)

	c, err := s.GRPC.PutEntry(ctx)
	if err != nil {
		return nil, err
	}

	chunk := make([]byte, chunkSize)
	bar := progressbar.DefaultBytes(int64(size))
	for {
		n, err := r.Read(chunk)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		req := pb.PutEntryRequest{
			Chunk: &pb.Chunk{
				Data: chunk[:n],
			},
		}
		err = c.Send(&req)
		if err != nil {
			return nil, err
		}
		bar.Add(n)
	}
	bar.Finish()
	log.Debugf(ctx, "done sending chunks")

	res, err := c.CloseAndRecv()
	if err != nil {
		return nil, err
	}
	if len(res.GetMetadata().GetDigests()) != 1 {
		return utils.Digest{}, fmt.Errorf("expected 1 digest, got %d", len(res.GetMetadata().GetDigests()))
	}
	log.Infof(ctx, "put entry: %v", res)

	digest := utils.DigestFromProto(res.Metadata.Digests[0])

	return digest, nil
}

func (s Remote) Has(ctx context.Context, digest utils.Digest) (bool, error) {
	log.Debugf(ctx, "checking existence of %s", utils.DigestForLog(digest))
	md := metadata.New(nil)
	md.Set(APIKeyHeader, s.APIKey)
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := pb.GetEntryMetadataRequest{
		Digest: utils.DigestToProto(digest),
	}
	res, err := s.GRPC.GetEntryMetadata(ctx, &req)
	if grpc.Code(err) == codes.NotFound {
		log.Debugf(ctx, "entry not found: %s", err)
		return false, nil
	} else if err != nil {
		log.Debugf(ctx, "error checking entry existence: %s", err)
		return false, err
	}

	ok := len(res.GetMetadata().GetDigests()) > 0
	log.Debugf(ctx, "entry found: %v", res)

	return ok, nil
}
