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

package main

import (
	"bytes"
	"context"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/ent/log"
	pb "github.com/google/ent/proto"
	"github.com/google/ent/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func redact(s string) string {
	if len(s) > 4 {
		return s[:4] + "..."
	}
	return s
}

type grpcServer struct {
	pb.UnimplementedEntServer
}

var _ pb.EntServer = grpcServer{}

// GetEntry implements ent.EntServer
func (grpcServer) GetEntry(req *pb.GetEntryRequest, s pb.Ent_GetEntryServer) error {
	ctx := s.Context()
	log.Infof(ctx, "GetEntry req: %s", req)
	accessItem := &LogItemGet{
		// TODO
		Source: SourceAPI,
	}
	defer LogGet(ctx, accessItem)

	apiKey := getAPIKeyGRPC(ctx)
	log.Debugf(ctx, "apiKey: %q", redact(apiKey))
	user := apiKeyToUser[apiKey]
	if user == nil {
		log.Warningf(ctx, "invalid API key: %q", redact(apiKey))
		return status.Errorf(codes.PermissionDenied, "invalid API key: %q", redact(apiKey))
	}
	log.Debugf(ctx, "user: %q %d", user.Name, user.ID)
	log.Debugf(ctx, "perms: read:%v write:%v", user.CanRead, user.CanWrite)
	if !user.CanRead {
		log.Warningf(ctx, "user %d does not have read permission", user.ID)
		return status.Errorf(codes.PermissionDenied, "user %d does not have read permission", user.ID)
	}
	accessItem.UserID = int64(user.ID)

	digest := utils.DigestFromProto(req.Digest)
	log.Debugf(ctx, "digest: %q", digest.String())

	log.Debugf(ctx, "getting blob: %q", digest.String())
	blob, err := blobStore.Get(ctx, digest)
	if err == storage.ErrObjectNotExist {
		log.Warningf(ctx, "blob not found: %q", digest.String())
		return status.Errorf(codes.NotFound, "blob not found: %q", digest.String())
	} else if err != nil {
		log.Warningf(ctx, "could not get blob: %s", err)
		return status.Errorf(codes.Internal, "could not get blob: %s", err)
	}
	log.Debugf(ctx, "got blob: %q", digest.String())

	err = s.Send(&pb.GetEntryResponse{
		Entry: &pb.GetEntryResponse_Metadata{
			Metadata: &pb.EntryMetadata{
				Digests: []*pb.Digest{
					utils.DigestToProto(digest),
				},
			},
		},
	})
	if err != nil {
		log.Warningf(ctx, "could not send response: %s", err)
		return status.Errorf(codes.Internal, "could not send response: %s", err)
	}

	err = s.Send(&pb.GetEntryResponse{
		Entry: &pb.GetEntryResponse_Chunk{
			Chunk: &pb.Chunk{
				Data: blob,
			},
		},
	})
	if err != nil {
		log.Warningf(ctx, "could not send response: %s", err)
		return status.Errorf(codes.Internal, "could not send response: %s", err)
	}

	return nil
}

func (grpcServer) GetEntryMetadata(ctx context.Context, req *pb.GetEntryMetadataRequest) (*pb.GetEntryMetadataResponse, error) {
	log.Infof(ctx, "HasEntry req: %s", req)
	accessItem := &LogItemGet{
		// TODO
		Source: SourceAPI,
	}
	defer LogGet(ctx, accessItem)

	apiKey := getAPIKeyGRPC(ctx)
	log.Debugf(ctx, "apiKey: %q", redact(apiKey))
	user := apiKeyToUser[apiKey]
	if user == nil {
		log.Warningf(ctx, "invalid API key: %q", redact(apiKey))
		return nil, status.Errorf(codes.PermissionDenied, "invalid API key: %q", redact(apiKey))
	}
	log.Debugf(ctx, "user: %q %d", user.Name, user.ID)
	log.Debugf(ctx, "perms: read:%v write:%v", user.CanRead, user.CanWrite)
	if !user.CanRead {
		log.Warningf(ctx, "user %d does not have read permission", user.ID)
		return nil, status.Errorf(codes.PermissionDenied, "user %d does not have read permission", user.ID)
	}
	accessItem.UserID = int64(user.ID)

	digest := utils.DigestFromProto(req.Digest)
	log.Debugf(ctx, "digest: %q", digest.String())

	log.Debugf(ctx, "getting blob: %q", digest.String())
	ok, err := blobStore.Has(ctx, digest)
	if err != nil {
		log.Warningf(ctx, "could not get blob: %s", err)
		return nil, status.Errorf(codes.Internal, "could not get blob: %s", err)
	}
	log.Debugf(ctx, "got blob: %q = %v", digest.String(), ok)

	if !ok {
		return nil, status.Errorf(codes.NotFound, "blob not found: %q", digest.String())
	}

	res := &pb.GetEntryMetadataResponse{
		Metadata: &pb.EntryMetadata{
			Digests: []*pb.Digest{
				utils.DigestToProto(digest),
			},
		},
	}

	return res, nil
}

// PutEntry implements ent.EntServer
func (grpcServer) PutEntry(s pb.Ent_PutEntryServer) error {
	ctx := s.Context()
	accessItem := &LogItemPut{
		// TODO
		Source: SourceAPI,
	}
	defer LogPut(ctx, accessItem)

	apiKey := getAPIKeyGRPC(ctx)
	log.Debugf(ctx, "apiKey: %q", redact(apiKey))
	user := apiKeyToUser[apiKey]
	if user == nil {
		log.Warningf(ctx, "invalid API key: %q", redact(apiKey))
		return status.Errorf(codes.PermissionDenied, "invalid API key: %q", redact(apiKey))
	}
	log.Debugf(ctx, "user: %q %d", user.Name, user.ID)
	log.Debugf(ctx, "perms: read:%v write:%v", user.CanRead, user.CanWrite)
	if !user.CanWrite {
		log.Warningf(ctx, "user %d does not have write permission", user.ID)
		return status.Errorf(codes.PermissionDenied, "user %d does not have write permission", user.ID)
	}
	accessItem.UserID = int64(user.ID)

	// TODO: Use correct size.
	blob := make([]byte, 0, 1024*1024)

	next := true
	for next {
		req, err := s.Recv()
		if err == io.EOF {
			log.Infof(ctx, "received EOF")
			blob = append(blob, req.GetChunk().GetData()...)
			next = false
		} else if err != nil {
			log.Warningf(ctx, "could not receive request: %s", err)
			return status.Errorf(codes.Internal, "could not receive request: %s", err)
		} else {
			blob = append(blob, req.GetChunk().GetData()...)
		}
	}

	digest := utils.ComputeDigest(blob)

	exists, err := blobStore.Has(ctx, digest)
	if err != nil {
		log.Errorf(ctx, "error checking blob existence: %s", err)
		accessItem.NotCreated = append(accessItem.NotCreated, digest.String())
	}
	if exists {
		log.Infof(ctx, "blob %q already exists", digest)
		accessItem.NotCreated = append(accessItem.NotCreated, digest.String())
		// We count the blob as created, even though it already existed.
	} else {
		log.Infof(ctx, "adding blob: %q", digest)
		digest1, err := blobStore.Put(ctx, blob)
		if !bytes.Equal(digest1, digest) {
			log.Errorf(ctx, "mismatching digest, expected %q, got %q", digest.String(), digest1.String())
		}
		accessItem.Digest = append(accessItem.Digest, digest1.String())
		if err != nil {
			log.Errorf(ctx, "error adding blob: %s", err)
			accessItem.NotCreated = append(accessItem.NotCreated, digest1.String())
		}
		log.Infof(ctx, "added blob: %q", digest1.String())
		accessItem.Created = append(accessItem.Created, digest1.String())
	}
	res := &pb.PutEntryResponse{
		Metadata: &pb.EntryMetadata{
			Digests: []*pb.Digest{
				utils.DigestToProto(digest),
			},
		},
	}
	err = s.SendAndClose(res)
	if err != nil {
		log.Warningf(ctx, "could not send response: %s", err)
		return status.Errorf(codes.Internal, "could not send response: %s", err)
	}
	return nil
}

// GetTag implements ent.EntServer
func (grpcServer) GetTag(ctx context.Context, req *pb.GetTagRequest) (*pb.GetTagResponse, error) {
	log.Debugf(ctx, "req: %s", req)

	entry, err := store.GetMapEntry(ctx, req.PublicKey, req.Label)
	if err != nil {
		log.Errorf(ctx, "could not get tag: %s", err)
		return nil, status.Errorf(codes.Internal, "could not get tag: %s", err)
	}
	if entry == nil {
		log.Debugf(ctx, "tag not found")
		return &pb.GetTagResponse{}, nil
	}
	log.Infof(ctx, "tag found: %v", entry)
	return &pb.GetTagResponse{
		SignedTag: &pb.SignedTag{
			Tag: &pb.Tag{
				Label: entry.Label,
				Target: &pb.Digest{
					Code:   uint64(entry.Target.Code),
					Digest: entry.Target.Digest,
				},
			},
		},
	}, nil

}

// SetTag implements ent.EntServer
func (grpcServer) SetTag(ctx context.Context, req *pb.SetTagRequest) (*pb.SetTagResponse, error) {
	log.Debugf(ctx, "req: %s", req)

	e := MapEntry{
		PublicKey: req.SignedTag.PublicKey,
		Label:     req.SignedTag.Tag.Label,
		Target: Digest{
			Code:   int64(req.SignedTag.Tag.Target.Code),
			Digest: req.SignedTag.Tag.Target.Digest,
		},
		EntrySignature: req.SignedTag.TagSignature,
		CreationTime:   time.Now(),
	}
	// TODO: Check signature.
	err := store.SetMapEntry(ctx, &e)
	if err != nil {
		log.Errorf(ctx, "could not set tag: %s", err)
		return nil, status.Errorf(codes.Internal, "could not set tag: %s", err)
	}
	log.Infof(ctx, "set tag: %v", e)

	return &pb.SetTagResponse{}, nil
}
