// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: steady.proto

package rpc

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

type CreateServiceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *CreateServiceRequest) Reset() {
	*x = CreateServiceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_steady_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateServiceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateServiceRequest) ProtoMessage() {}

func (x *CreateServiceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_steady_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateServiceRequest.ProtoReflect.Descriptor instead.
func (*CreateServiceRequest) Descriptor() ([]byte, []int) {
	return file_steady_proto_rawDescGZIP(), []int{0}
}

func (x *CreateServiceRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type CreateServiceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Service *Service `protobuf:"bytes,1,opt,name=service,proto3" json:"service,omitempty"`
}

func (x *CreateServiceResponse) Reset() {
	*x = CreateServiceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_steady_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateServiceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateServiceResponse) ProtoMessage() {}

func (x *CreateServiceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_steady_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateServiceResponse.ProtoReflect.Descriptor instead.
func (*CreateServiceResponse) Descriptor() ([]byte, []int) {
	return file_steady_proto_rawDescGZIP(), []int{1}
}

func (x *CreateServiceResponse) GetService() *Service {
	if x != nil {
		return x.Service
	}
	return nil
}

type Service struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name   string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Id     int64  `protobuf:"varint,2,opt,name=id,proto3" json:"id,omitempty"`
	UserId int64  `protobuf:"varint,3,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

func (x *Service) Reset() {
	*x = Service{}
	if protoimpl.UnsafeEnabled {
		mi := &file_steady_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Service) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Service) ProtoMessage() {}

func (x *Service) ProtoReflect() protoreflect.Message {
	mi := &file_steady_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Service.ProtoReflect.Descriptor instead.
func (*Service) Descriptor() ([]byte, []int) {
	return file_steady_proto_rawDescGZIP(), []int{2}
}

func (x *Service) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Service) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Service) GetUserId() int64 {
	if x != nil {
		return x.UserId
	}
	return 0
}

type CreateServiceVersionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceId int64  `protobuf:"varint,1,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
	Version   string `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Source    string `protobuf:"bytes,3,opt,name=source,proto3" json:"source,omitempty"`
}

func (x *CreateServiceVersionRequest) Reset() {
	*x = CreateServiceVersionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_steady_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateServiceVersionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateServiceVersionRequest) ProtoMessage() {}

func (x *CreateServiceVersionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_steady_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateServiceVersionRequest.ProtoReflect.Descriptor instead.
func (*CreateServiceVersionRequest) Descriptor() ([]byte, []int) {
	return file_steady_proto_rawDescGZIP(), []int{3}
}

func (x *CreateServiceVersionRequest) GetServiceId() int64 {
	if x != nil {
		return x.ServiceId
	}
	return 0
}

func (x *CreateServiceVersionRequest) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *CreateServiceVersionRequest) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

type CreateServiceVersionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceVersion *ServiceVersion `protobuf:"bytes,1,opt,name=service_version,json=serviceVersion,proto3" json:"service_version,omitempty"`
}

func (x *CreateServiceVersionResponse) Reset() {
	*x = CreateServiceVersionResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_steady_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateServiceVersionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateServiceVersionResponse) ProtoMessage() {}

func (x *CreateServiceVersionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_steady_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateServiceVersionResponse.ProtoReflect.Descriptor instead.
func (*CreateServiceVersionResponse) Descriptor() ([]byte, []int) {
	return file_steady_proto_rawDescGZIP(), []int{4}
}

func (x *CreateServiceVersionResponse) GetServiceVersion() *ServiceVersion {
	if x != nil {
		return x.ServiceVersion
	}
	return nil
}

type ServiceVersion struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        int64  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	ServiceId int64  `protobuf:"varint,2,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
	Version   string `protobuf:"bytes,3,opt,name=version,proto3" json:"version,omitempty"`
	Source    string `protobuf:"bytes,4,opt,name=source,proto3" json:"source,omitempty"`
}

func (x *ServiceVersion) Reset() {
	*x = ServiceVersion{}
	if protoimpl.UnsafeEnabled {
		mi := &file_steady_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceVersion) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceVersion) ProtoMessage() {}

func (x *ServiceVersion) ProtoReflect() protoreflect.Message {
	mi := &file_steady_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceVersion.ProtoReflect.Descriptor instead.
func (*ServiceVersion) Descriptor() ([]byte, []int) {
	return file_steady_proto_rawDescGZIP(), []int{5}
}

func (x *ServiceVersion) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *ServiceVersion) GetServiceId() int64 {
	if x != nil {
		return x.ServiceId
	}
	return 0
}

func (x *ServiceVersion) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *ServiceVersion) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

type DeployApplicationRequeast struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceVersionId int64  `protobuf:"varint,1,opt,name=service_version_id,json=serviceVersionId,proto3" json:"service_version_id,omitempty"`
	Name             string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *DeployApplicationRequeast) Reset() {
	*x = DeployApplicationRequeast{}
	if protoimpl.UnsafeEnabled {
		mi := &file_steady_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeployApplicationRequeast) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeployApplicationRequeast) ProtoMessage() {}

func (x *DeployApplicationRequeast) ProtoReflect() protoreflect.Message {
	mi := &file_steady_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeployApplicationRequeast.ProtoReflect.Descriptor instead.
func (*DeployApplicationRequeast) Descriptor() ([]byte, []int) {
	return file_steady_proto_rawDescGZIP(), []int{6}
}

func (x *DeployApplicationRequeast) GetServiceVersionId() int64 {
	if x != nil {
		return x.ServiceVersionId
	}
	return 0
}

func (x *DeployApplicationRequeast) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type DeployApplicationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Application *Application `protobuf:"bytes,1,opt,name=application,proto3" json:"application,omitempty"`
	Url         string       `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
}

func (x *DeployApplicationResponse) Reset() {
	*x = DeployApplicationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_steady_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeployApplicationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeployApplicationResponse) ProtoMessage() {}

func (x *DeployApplicationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_steady_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeployApplicationResponse.ProtoReflect.Descriptor instead.
func (*DeployApplicationResponse) Descriptor() ([]byte, []int) {
	return file_steady_proto_rawDescGZIP(), []int{7}
}

func (x *DeployApplicationResponse) GetApplication() *Application {
	if x != nil {
		return x.Application
	}
	return nil
}

func (x *DeployApplicationResponse) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

type Application struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id               int64  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	ServiceVersionId int64  `protobuf:"varint,2,opt,name=service_version_id,json=serviceVersionId,proto3" json:"service_version_id,omitempty"`
	UserId           int64  `protobuf:"varint,3,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Name             string `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *Application) Reset() {
	*x = Application{}
	if protoimpl.UnsafeEnabled {
		mi := &file_steady_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Application) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Application) ProtoMessage() {}

func (x *Application) ProtoReflect() protoreflect.Message {
	mi := &file_steady_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Application.ProtoReflect.Descriptor instead.
func (*Application) Descriptor() ([]byte, []int) {
	return file_steady_proto_rawDescGZIP(), []int{8}
}

func (x *Application) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Application) GetServiceVersionId() int64 {
	if x != nil {
		return x.ServiceVersionId
	}
	return 0
}

func (x *Application) GetUserId() int64 {
	if x != nil {
		return x.UserId
	}
	return 0
}

func (x *Application) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type DeploySourceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Source string `protobuf:"bytes,1,opt,name=source,proto3" json:"source,omitempty"`
}

func (x *DeploySourceRequest) Reset() {
	*x = DeploySourceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_steady_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeploySourceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeploySourceRequest) ProtoMessage() {}

func (x *DeploySourceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_steady_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeploySourceRequest.ProtoReflect.Descriptor instead.
func (*DeploySourceRequest) Descriptor() ([]byte, []int) {
	return file_steady_proto_rawDescGZIP(), []int{9}
}

func (x *DeploySourceRequest) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

type DeploySourceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Url string `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
}

func (x *DeploySourceResponse) Reset() {
	*x = DeploySourceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_steady_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeploySourceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeploySourceResponse) ProtoMessage() {}

func (x *DeploySourceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_steady_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeploySourceResponse.ProtoReflect.Descriptor instead.
func (*DeploySourceResponse) Descriptor() ([]byte, []int) {
	return file_steady_proto_rawDescGZIP(), []int{10}
}

func (x *DeploySourceResponse) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

var File_steady_proto protoreflect.FileDescriptor

var file_steady_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d,
	0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x22, 0x2a, 0x0a,
	0x14, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x49, 0x0a, 0x15, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x30, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x73, 0x74, 0x65,
	0x61, 0x64, 0x79, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x07, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x22, 0x46, 0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x22, 0x6e, 0x0a, 0x1b,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x09, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x22, 0x66, 0x0a, 0x1c,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x46, 0x0a, 0x0f,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x73,
	0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x22, 0x71, 0x0a, 0x0e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12,
	0x16, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x22, 0x5d, 0x0a, 0x19, 0x44, 0x65, 0x70, 0x6c, 0x6f,
	0x79, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x61, 0x73, 0x74, 0x12, 0x2c, 0x0a, 0x12, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x10, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x6b, 0x0a, 0x19, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79,
	0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x3c, 0x0a, 0x0b, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64,
	0x79, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x75, 0x72, 0x6c, 0x22, 0x78, 0x0a, 0x0b, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x2c, 0x0a, 0x12, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x10,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x64,
	0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x2d, 0x0a,
	0x13, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x22, 0x28, 0x0a, 0x14,
	0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x32, 0x97, 0x03, 0x0a, 0x06, 0x53, 0x74, 0x65, 0x61, 0x64,
	0x79, 0x12, 0x5a, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x23, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x73, 0x74, 0x65, 0x61,
	0x64, 0x79, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79,
	0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x6f, 0x0a,
	0x14, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x2a, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x73,
	0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x2b, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64,
	0x79, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x67,
	0x0a, 0x11, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x28, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x73, 0x74, 0x65,
	0x61, 0x64, 0x79, 0x2e, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x61, 0x73, 0x74, 0x1a, 0x28, 0x2e,
	0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x44, 0x65,
	0x70, 0x6c, 0x6f, 0x79, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x57, 0x0a, 0x0c, 0x44, 0x65, 0x70, 0x6c, 0x6f,
	0x79, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x22, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79,
	0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x53, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x73, 0x74,
	0x65, 0x61, 0x64, 0x79, 0x2e, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2e, 0x44, 0x65, 0x70, 0x6c,
	0x6f, 0x79, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x42, 0x25, 0x5a, 0x23, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d,
	0x61, 0x78, 0x6d, 0x63, 0x64, 0x2f, 0x73, 0x74, 0x65, 0x61, 0x64, 0x79, 0x2f, 0x73, 0x74, 0x65,
	0x61, 0x64, 0x79, 0x2f, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_steady_proto_rawDescOnce sync.Once
	file_steady_proto_rawDescData = file_steady_proto_rawDesc
)

func file_steady_proto_rawDescGZIP() []byte {
	file_steady_proto_rawDescOnce.Do(func() {
		file_steady_proto_rawDescData = protoimpl.X.CompressGZIP(file_steady_proto_rawDescData)
	})
	return file_steady_proto_rawDescData
}

var file_steady_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_steady_proto_goTypes = []interface{}{
	(*CreateServiceRequest)(nil),         // 0: steady.steady.CreateServiceRequest
	(*CreateServiceResponse)(nil),        // 1: steady.steady.CreateServiceResponse
	(*Service)(nil),                      // 2: steady.steady.Service
	(*CreateServiceVersionRequest)(nil),  // 3: steady.steady.CreateServiceVersionRequest
	(*CreateServiceVersionResponse)(nil), // 4: steady.steady.CreateServiceVersionResponse
	(*ServiceVersion)(nil),               // 5: steady.steady.ServiceVersion
	(*DeployApplicationRequeast)(nil),    // 6: steady.steady.DeployApplicationRequeast
	(*DeployApplicationResponse)(nil),    // 7: steady.steady.DeployApplicationResponse
	(*Application)(nil),                  // 8: steady.steady.Application
	(*DeploySourceRequest)(nil),          // 9: steady.steady.DeploySourceRequest
	(*DeploySourceResponse)(nil),         // 10: steady.steady.DeploySourceResponse
}
var file_steady_proto_depIdxs = []int32{
	2,  // 0: steady.steady.CreateServiceResponse.service:type_name -> steady.steady.Service
	5,  // 1: steady.steady.CreateServiceVersionResponse.service_version:type_name -> steady.steady.ServiceVersion
	8,  // 2: steady.steady.DeployApplicationResponse.application:type_name -> steady.steady.Application
	0,  // 3: steady.steady.Steady.CreateService:input_type -> steady.steady.CreateServiceRequest
	3,  // 4: steady.steady.Steady.CreateServiceVersion:input_type -> steady.steady.CreateServiceVersionRequest
	6,  // 5: steady.steady.Steady.DeployApplication:input_type -> steady.steady.DeployApplicationRequeast
	9,  // 6: steady.steady.Steady.DeploySource:input_type -> steady.steady.DeploySourceRequest
	1,  // 7: steady.steady.Steady.CreateService:output_type -> steady.steady.CreateServiceResponse
	4,  // 8: steady.steady.Steady.CreateServiceVersion:output_type -> steady.steady.CreateServiceVersionResponse
	7,  // 9: steady.steady.Steady.DeployApplication:output_type -> steady.steady.DeployApplicationResponse
	10, // 10: steady.steady.Steady.DeploySource:output_type -> steady.steady.DeploySourceResponse
	7,  // [7:11] is the sub-list for method output_type
	3,  // [3:7] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_steady_proto_init() }
func file_steady_proto_init() {
	if File_steady_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_steady_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateServiceRequest); i {
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
		file_steady_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateServiceResponse); i {
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
		file_steady_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Service); i {
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
		file_steady_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateServiceVersionRequest); i {
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
		file_steady_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateServiceVersionResponse); i {
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
		file_steady_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceVersion); i {
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
		file_steady_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeployApplicationRequeast); i {
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
		file_steady_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeployApplicationResponse); i {
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
		file_steady_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Application); i {
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
		file_steady_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeploySourceRequest); i {
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
		file_steady_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeploySourceResponse); i {
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
			RawDescriptor: file_steady_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_steady_proto_goTypes,
		DependencyIndexes: file_steady_proto_depIdxs,
		MessageInfos:      file_steady_proto_msgTypes,
	}.Build()
	File_steady_proto = out.File
	file_steady_proto_rawDesc = nil
	file_steady_proto_goTypes = nil
	file_steady_proto_depIdxs = nil
}
