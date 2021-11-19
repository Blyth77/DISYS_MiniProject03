//
//Generate-files:
//protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ChittyChat.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: auctionhouse.proto

package proto

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

type Status int32

const (
	Status_INVALID Status = 0
	Status_SUCCESS Status = 1
	Status_FAILURE Status = 2
)

// Enum value maps for Status.
var (
	Status_name = map[int32]string{
		0: "INVALID",
		1: "SUCCESS",
		2: "FAILURE",
	}
	Status_value = map[string]int32{
		"INVALID": 0,
		"SUCCESS": 1,
		"FAILURE": 2,
	}
)

func (x Status) Enum() *Status {
	p := new(Status)
	*p = x
	return p
}

func (x Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status) Descriptor() protoreflect.EnumDescriptor {
	return file_auctionhouse_proto_enumTypes[0].Descriptor()
}

func (Status) Type() protoreflect.EnumType {
	return &file_auctionhouse_proto_enumTypes[0]
}

func (x Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status.Descriptor instead.
func (Status) EnumDescriptor() ([]byte, []int) {
	return file_auctionhouse_proto_rawDescGZIP(), []int{0}
}

type Subscription struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientId         int32  `protobuf:"varint,1,opt,name=clientId,proto3" json:"clientId,omitempty"`
	UserName         string `protobuf:"bytes,2,opt,name=userName,proto3" json:"userName,omitempty"`
	LamportTimestamp int32  `protobuf:"varint,3,opt,name=lamportTimestamp,proto3" json:"lamportTimestamp,omitempty"`
}

func (x *Subscription) Reset() {
	*x = Subscription{}
	if protoimpl.UnsafeEnabled {
		mi := &file_auctionhouse_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Subscription) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Subscription) ProtoMessage() {}

func (x *Subscription) ProtoReflect() protoreflect.Message {
	mi := &file_auctionhouse_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Subscription.ProtoReflect.Descriptor instead.
func (*Subscription) Descriptor() ([]byte, []int) {
	return file_auctionhouse_proto_rawDescGZIP(), []int{0}
}

func (x *Subscription) GetClientId() int32 {
	if x != nil {
		return x.ClientId
	}
	return 0
}

func (x *Subscription) GetUserName() string {
	if x != nil {
		return x.UserName
	}
	return ""
}

func (x *Subscription) GetLamportTimestamp() int32 {
	if x != nil {
		return x.LamportTimestamp
	}
	return 0
}

type ClientMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientId         int32  `protobuf:"varint,1,opt,name=clientId,proto3" json:"clientId,omitempty"`
	UserName         string `protobuf:"bytes,2,opt,name=userName,proto3" json:"userName,omitempty"`
	Msg              string `protobuf:"bytes,3,opt,name=msg,proto3" json:"msg,omitempty"`
	LamportTimestamp int32  `protobuf:"varint,4,opt,name=lamportTimestamp,proto3" json:"lamportTimestamp,omitempty"`
	Code             int32  `protobuf:"varint,5,opt,name=code,proto3" json:"code,omitempty"`
}

func (x *ClientMessage) Reset() {
	*x = ClientMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_auctionhouse_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientMessage) ProtoMessage() {}

func (x *ClientMessage) ProtoReflect() protoreflect.Message {
	mi := &file_auctionhouse_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientMessage.ProtoReflect.Descriptor instead.
func (*ClientMessage) Descriptor() ([]byte, []int) {
	return file_auctionhouse_proto_rawDescGZIP(), []int{1}
}

func (x *ClientMessage) GetClientId() int32 {
	if x != nil {
		return x.ClientId
	}
	return 0
}

func (x *ClientMessage) GetUserName() string {
	if x != nil {
		return x.UserName
	}
	return ""
}

func (x *ClientMessage) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

func (x *ClientMessage) GetLamportTimestamp() int32 {
	if x != nil {
		return x.LamportTimestamp
	}
	return 0
}

func (x *ClientMessage) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

type ChatRoomMessages struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msg              string `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	LamportTimestamp int32  `protobuf:"varint,2,opt,name=lamportTimestamp,proto3" json:"lamportTimestamp,omitempty"`
	Username         string `protobuf:"bytes,3,opt,name=username,proto3" json:"username,omitempty"`
	ClientId         int32  `protobuf:"varint,4,opt,name=clientId,proto3" json:"clientId,omitempty"`
	Code             int32  `protobuf:"varint,5,opt,name=code,proto3" json:"code,omitempty"`
}

func (x *ChatRoomMessages) Reset() {
	*x = ChatRoomMessages{}
	if protoimpl.UnsafeEnabled {
		mi := &file_auctionhouse_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChatRoomMessages) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatRoomMessages) ProtoMessage() {}

func (x *ChatRoomMessages) ProtoReflect() protoreflect.Message {
	mi := &file_auctionhouse_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChatRoomMessages.ProtoReflect.Descriptor instead.
func (*ChatRoomMessages) Descriptor() ([]byte, []int) {
	return file_auctionhouse_proto_rawDescGZIP(), []int{2}
}

func (x *ChatRoomMessages) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

func (x *ChatRoomMessages) GetLamportTimestamp() int32 {
	if x != nil {
		return x.LamportTimestamp
	}
	return 0
}

func (x *ChatRoomMessages) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *ChatRoomMessages) GetClientId() int32 {
	if x != nil {
		return x.ClientId
	}
	return 0
}

func (x *ChatRoomMessages) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

type StatusMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Operation string `protobuf:"bytes,1,opt,name=operation,proto3" json:"operation,omitempty"`
	Status    Status `protobuf:"varint,2,opt,name=status,proto3,enum=proto.Status" json:"status,omitempty"`
}

func (x *StatusMessage) Reset() {
	*x = StatusMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_auctionhouse_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatusMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusMessage) ProtoMessage() {}

func (x *StatusMessage) ProtoReflect() protoreflect.Message {
	mi := &file_auctionhouse_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusMessage.ProtoReflect.Descriptor instead.
func (*StatusMessage) Descriptor() ([]byte, []int) {
	return file_auctionhouse_proto_rawDescGZIP(), []int{3}
}

func (x *StatusMessage) GetOperation() string {
	if x != nil {
		return x.Operation
	}
	return ""
}

func (x *StatusMessage) GetStatus() Status {
	if x != nil {
		return x.Status
	}
	return Status_INVALID
}

var File_auctionhouse_proto protoreflect.FileDescriptor

var file_auctionhouse_proto_rawDesc = []byte{
	0x0a, 0x12, 0x61, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x68, 0x6f, 0x75, 0x73, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x72, 0x0a, 0x0c, 0x53,
	0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x63,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x63,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x4e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x2a, 0x0a, 0x10, 0x6c, 0x61, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x10, 0x6c,
	0x61, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x22,
	0x99, 0x01, 0x0a, 0x0d, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x08, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a,
	0x08, 0x75, 0x73, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x75, 0x73, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x12, 0x2a, 0x0a, 0x10, 0x6c,
	0x61, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x10, 0x6c, 0x61, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x22, 0x9c, 0x01, 0x0a, 0x10,
	0x43, 0x68, 0x61, 0x74, 0x52, 0x6f, 0x6f, 0x6d, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73,
	0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d,
	0x73, 0x67, 0x12, 0x2a, 0x0a, 0x10, 0x6c, 0x61, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x10, 0x6c, 0x61,
	0x6d, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x1a,
	0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x22, 0x54, 0x0a, 0x0d, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x6f,
	0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x06, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x2a, 0x2f, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0b, 0x0a, 0x07, 0x49, 0x4e,
	0x56, 0x41, 0x4c, 0x49, 0x44, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x55, 0x43, 0x43, 0x45,
	0x53, 0x53, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x46, 0x41, 0x49, 0x4c, 0x55, 0x52, 0x45, 0x10,
	0x02, 0x32, 0x8f, 0x01, 0x0a, 0x11, 0x43, 0x68, 0x69, 0x74, 0x74, 0x79, 0x43, 0x68, 0x61, 0x74,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3b, 0x0a, 0x07, 0x50, 0x75, 0x62, 0x6c, 0x69,
	0x73, 0x68, 0x12, 0x14, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x14, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x00,
	0x28, 0x01, 0x30, 0x01, 0x12, 0x3d, 0x0a, 0x09, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73,
	0x74, 0x12, 0x13, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43,
	0x68, 0x61, 0x74, 0x52, 0x6f, 0x6f, 0x6d, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x22,
	0x00, 0x30, 0x01, 0x42, 0x09, 0x5a, 0x07, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_auctionhouse_proto_rawDescOnce sync.Once
	file_auctionhouse_proto_rawDescData = file_auctionhouse_proto_rawDesc
)

func file_auctionhouse_proto_rawDescGZIP() []byte {
	file_auctionhouse_proto_rawDescOnce.Do(func() {
		file_auctionhouse_proto_rawDescData = protoimpl.X.CompressGZIP(file_auctionhouse_proto_rawDescData)
	})
	return file_auctionhouse_proto_rawDescData
}

var file_auctionhouse_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_auctionhouse_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_auctionhouse_proto_goTypes = []interface{}{
	(Status)(0),              // 0: proto.Status
	(*Subscription)(nil),     // 1: proto.Subscription
	(*ClientMessage)(nil),    // 2: proto.ClientMessage
	(*ChatRoomMessages)(nil), // 3: proto.ChatRoomMessages
	(*StatusMessage)(nil),    // 4: proto.StatusMessage
}
var file_auctionhouse_proto_depIdxs = []int32{
	0, // 0: proto.StatusMessage.status:type_name -> proto.Status
	2, // 1: proto.ChittyChatService.Publish:input_type -> proto.ClientMessage
	1, // 2: proto.ChittyChatService.Broadcast:input_type -> proto.Subscription
	4, // 3: proto.ChittyChatService.Publish:output_type -> proto.StatusMessage
	3, // 4: proto.ChittyChatService.Broadcast:output_type -> proto.ChatRoomMessages
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_auctionhouse_proto_init() }
func file_auctionhouse_proto_init() {
	if File_auctionhouse_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_auctionhouse_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Subscription); i {
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
		file_auctionhouse_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientMessage); i {
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
		file_auctionhouse_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChatRoomMessages); i {
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
		file_auctionhouse_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatusMessage); i {
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
			RawDescriptor: file_auctionhouse_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_auctionhouse_proto_goTypes,
		DependencyIndexes: file_auctionhouse_proto_depIdxs,
		EnumInfos:         file_auctionhouse_proto_enumTypes,
		MessageInfos:      file_auctionhouse_proto_msgTypes,
	}.Build()
	File_auctionhouse_proto = out.File
	file_auctionhouse_proto_rawDesc = nil
	file_auctionhouse_proto_goTypes = nil
	file_auctionhouse_proto_depIdxs = nil
}
