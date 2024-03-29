// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.8
// source: proto/ent_server_config.proto

package ent

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProjectId     string        `protobuf:"bytes,1,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	ListenAddress string        `protobuf:"bytes,2,opt,name=listen_address,json=listenAddress,proto3" json:"listen_address,omitempty"`
	DomainName    string        `protobuf:"bytes,3,opt,name=domain_name,json=domainName,proto3" json:"domain_name,omitempty"`
	Redis         *Redis        `protobuf:"bytes,4,opt,name=redis,proto3" json:"redis,omitempty"`
	BigQuery      *BigQuery     `protobuf:"bytes,5,opt,name=big_query,json=bigQuery,proto3" json:"big_query,omitempty"`
	CloudStorage  *CloudStorage `protobuf:"bytes,6,opt,name=cloud_storage,json=cloudStorage,proto3" json:"cloud_storage,omitempty"`
	GinMode       string        `protobuf:"bytes,7,opt,name=gin_mode,json=ginMode,proto3" json:"gin_mode,omitempty"`
	LogLevel      string        `protobuf:"bytes,8,opt,name=log_level,json=logLevel,proto3" json:"log_level,omitempty"`
	Remote        []*Remote     `protobuf:"bytes,9,rep,name=remote,proto3" json:"remote,omitempty"`
	User          []*User       `protobuf:"bytes,10,rep,name=user,proto3" json:"user,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_ent_server_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_proto_ent_server_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_proto_ent_server_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

func (x *Config) GetListenAddress() string {
	if x != nil {
		return x.ListenAddress
	}
	return ""
}

func (x *Config) GetDomainName() string {
	if x != nil {
		return x.DomainName
	}
	return ""
}

func (x *Config) GetRedis() *Redis {
	if x != nil {
		return x.Redis
	}
	return nil
}

func (x *Config) GetBigQuery() *BigQuery {
	if x != nil {
		return x.BigQuery
	}
	return nil
}

func (x *Config) GetCloudStorage() *CloudStorage {
	if x != nil {
		return x.CloudStorage
	}
	return nil
}

func (x *Config) GetGinMode() string {
	if x != nil {
		return x.GinMode
	}
	return ""
}

func (x *Config) GetLogLevel() string {
	if x != nil {
		return x.LogLevel
	}
	return ""
}

func (x *Config) GetRemote() []*Remote {
	if x != nil {
		return x.Remote
	}
	return nil
}

func (x *Config) GetUser() []*User {
	if x != nil {
		return x.User
	}
	return nil
}

type Redis struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Endpoint string `protobuf:"bytes,1,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
}

func (x *Redis) Reset() {
	*x = Redis{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_ent_server_config_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Redis) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Redis) ProtoMessage() {}

func (x *Redis) ProtoReflect() protoreflect.Message {
	mi := &file_proto_ent_server_config_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Redis.ProtoReflect.Descriptor instead.
func (*Redis) Descriptor() ([]byte, []int) {
	return file_proto_ent_server_config_proto_rawDescGZIP(), []int{1}
}

func (x *Redis) GetEndpoint() string {
	if x != nil {
		return x.Endpoint
	}
	return ""
}

type BigQuery struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Dataset string `protobuf:"bytes,1,opt,name=dataset,proto3" json:"dataset,omitempty"`
}

func (x *BigQuery) Reset() {
	*x = BigQuery{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_ent_server_config_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BigQuery) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BigQuery) ProtoMessage() {}

func (x *BigQuery) ProtoReflect() protoreflect.Message {
	mi := &file_proto_ent_server_config_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BigQuery.ProtoReflect.Descriptor instead.
func (*BigQuery) Descriptor() ([]byte, []int) {
	return file_proto_ent_server_config_proto_rawDescGZIP(), []int{2}
}

func (x *BigQuery) GetDataset() string {
	if x != nil {
		return x.Dataset
	}
	return ""
}

type CloudStorage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Bucket string `protobuf:"bytes,1,opt,name=bucket,proto3" json:"bucket,omitempty"`
}

func (x *CloudStorage) Reset() {
	*x = CloudStorage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_ent_server_config_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CloudStorage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CloudStorage) ProtoMessage() {}

func (x *CloudStorage) ProtoReflect() protoreflect.Message {
	mi := &file_proto_ent_server_config_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CloudStorage.ProtoReflect.Descriptor instead.
func (*CloudStorage) Descriptor() ([]byte, []int) {
	return file_proto_ent_server_config_proto_rawDescGZIP(), []int{3}
}

func (x *CloudStorage) GetBucket() string {
	if x != nil {
		return x.Bucket
	}
	return ""
}

type Remote struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *Remote) Reset() {
	*x = Remote{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_ent_server_config_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Remote) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Remote) ProtoMessage() {}

func (x *Remote) ProtoReflect() protoreflect.Message {
	mi := &file_proto_ent_server_config_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Remote.ProtoReflect.Descriptor instead.
func (*Remote) Descriptor() ([]byte, []int) {
	return file_proto_ent_server_config_proto_rawDescGZIP(), []int{4}
}

func (x *Remote) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type User struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       int64  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name     string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	ApiKey   string `protobuf:"bytes,3,opt,name=api_key,json=apiKey,proto3" json:"api_key,omitempty"`
	CanRead  bool   `protobuf:"varint,4,opt,name=can_read,json=canRead,proto3" json:"can_read,omitempty"`
	CanWrite bool   `protobuf:"varint,5,opt,name=can_write,json=canWrite,proto3" json:"can_write,omitempty"`
}

func (x *User) Reset() {
	*x = User{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_ent_server_config_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *User) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*User) ProtoMessage() {}

func (x *User) ProtoReflect() protoreflect.Message {
	mi := &file_proto_ent_server_config_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use User.ProtoReflect.Descriptor instead.
func (*User) Descriptor() ([]byte, []int) {
	return file_proto_ent_server_config_proto_rawDescGZIP(), []int{5}
}

func (x *User) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *User) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *User) GetApiKey() string {
	if x != nil {
		return x.ApiKey
	}
	return ""
}

func (x *User) GetCanRead() bool {
	if x != nil {
		return x.CanRead
	}
	return false
}

func (x *User) GetCanWrite() bool {
	if x != nil {
		return x.CanWrite
	}
	return false
}

var File_proto_ent_server_config_proto protoreflect.FileDescriptor

var file_proto_ent_server_config_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x6e, 0x74, 0x5f, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x11, 0x65, 0x6e, 0x74, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x22, 0xb7, 0x03, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x1d, 0x0a,
	0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e,
	0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x2e, 0x0a, 0x05, 0x72, 0x65, 0x64, 0x69, 0x73, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x65, 0x6e, 0x74, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x52, 0x65, 0x64, 0x69, 0x73, 0x52, 0x05, 0x72,
	0x65, 0x64, 0x69, 0x73, 0x12, 0x38, 0x0a, 0x09, 0x62, 0x69, 0x67, 0x5f, 0x71, 0x75, 0x65, 0x72,
	0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x65, 0x6e, 0x74, 0x2e, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x42, 0x69, 0x67, 0x51,
	0x75, 0x65, 0x72, 0x79, 0x52, 0x08, 0x62, 0x69, 0x67, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x44,
	0x0a, 0x0d, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x5f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x65, 0x6e, 0x74, 0x2e, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x53,
	0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x52, 0x0c, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x53, 0x74, 0x6f,
	0x72, 0x61, 0x67, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x67, 0x69, 0x6e, 0x5f, 0x6d, 0x6f, 0x64, 0x65,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x67, 0x69, 0x6e, 0x4d, 0x6f, 0x64, 0x65, 0x12,
	0x1b, 0x0a, 0x09, 0x6c, 0x6f, 0x67, 0x5f, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x6c, 0x6f, 0x67, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x31, 0x0a, 0x06,
	0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x18, 0x09, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x65,
	0x6e, 0x74, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x52, 0x06, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x12,
	0x2b, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e,
	0x65, 0x6e, 0x74, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x52, 0x04, 0x75, 0x73, 0x65, 0x72, 0x22, 0x23, 0x0a, 0x05,
	0x52, 0x65, 0x64, 0x69, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x22, 0x24, 0x0a, 0x08, 0x42, 0x69, 0x67, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x18, 0x0a,
	0x07, 0x64, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x64, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x22, 0x26, 0x0a, 0x0c, 0x43, 0x6c, 0x6f, 0x75, 0x64,
	0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x22,
	0x1c, 0x0a, 0x06, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x7b, 0x0a,
	0x04, 0x55, 0x73, 0x65, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x61, 0x70, 0x69,
	0x5f, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x70, 0x69, 0x4b,
	0x65, 0x79, 0x12, 0x19, 0x0a, 0x08, 0x63, 0x61, 0x6e, 0x5f, 0x72, 0x65, 0x61, 0x64, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x63, 0x61, 0x6e, 0x52, 0x65, 0x61, 0x64, 0x12, 0x1b, 0x0a,
	0x09, 0x63, 0x61, 0x6e, 0x5f, 0x77, 0x72, 0x69, 0x74, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x08, 0x63, 0x61, 0x6e, 0x57, 0x72, 0x69, 0x74, 0x65, 0x42, 0x10, 0x5a, 0x0e, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x6e, 0x74, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_ent_server_config_proto_rawDescOnce sync.Once
	file_proto_ent_server_config_proto_rawDescData = file_proto_ent_server_config_proto_rawDesc
)

func file_proto_ent_server_config_proto_rawDescGZIP() []byte {
	file_proto_ent_server_config_proto_rawDescOnce.Do(func() {
		file_proto_ent_server_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_ent_server_config_proto_rawDescData)
	})
	return file_proto_ent_server_config_proto_rawDescData
}

var file_proto_ent_server_config_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_proto_ent_server_config_proto_goTypes = []interface{}{
	(*Config)(nil),       // 0: ent.server.config.Config
	(*Redis)(nil),        // 1: ent.server.config.Redis
	(*BigQuery)(nil),     // 2: ent.server.config.BigQuery
	(*CloudStorage)(nil), // 3: ent.server.config.CloudStorage
	(*Remote)(nil),       // 4: ent.server.config.Remote
	(*User)(nil),         // 5: ent.server.config.User
}
var file_proto_ent_server_config_proto_depIdxs = []int32{
	1, // 0: ent.server.config.Config.redis:type_name -> ent.server.config.Redis
	2, // 1: ent.server.config.Config.big_query:type_name -> ent.server.config.BigQuery
	3, // 2: ent.server.config.Config.cloud_storage:type_name -> ent.server.config.CloudStorage
	4, // 3: ent.server.config.Config.remote:type_name -> ent.server.config.Remote
	5, // 4: ent.server.config.Config.user:type_name -> ent.server.config.User
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_proto_ent_server_config_proto_init() }
func file_proto_ent_server_config_proto_init() {
	if File_proto_ent_server_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_ent_server_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_ent_server_config_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Redis); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_ent_server_config_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BigQuery); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_ent_server_config_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CloudStorage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_ent_server_config_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Remote); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_ent_server_config_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*User); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_ent_server_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_ent_server_config_proto_goTypes,
		DependencyIndexes: file_proto_ent_server_config_proto_depIdxs,
		MessageInfos:      file_proto_ent_server_config_proto_msgTypes,
	}.Build()
	File_proto_ent_server_config_proto = out.File
	file_proto_ent_server_config_proto_rawDesc = nil
	file_proto_ent_server_config_proto_goTypes = nil
	file_proto_ent_server_config_proto_depIdxs = nil
}
