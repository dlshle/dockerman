// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: proto/internal.proto

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

type ProxyHeader_Type int32

const (
	ProxyHeader_UNKONWN    ProxyHeader_Type = 0
	ProxyHeader_DATA       ProxyHeader_Type = 1
	ProxyHeader_CONNECT    ProxyHeader_Type = 2
	ProxyHeader_DISCONNECT ProxyHeader_Type = 3
	ProxyHeader_ACK        ProxyHeader_Type = 4
)

// Enum value maps for ProxyHeader_Type.
var (
	ProxyHeader_Type_name = map[int32]string{
		0: "UNKONWN",
		1: "DATA",
		2: "CONNECT",
		3: "DISCONNECT",
		4: "ACK",
	}
	ProxyHeader_Type_value = map[string]int32{
		"UNKONWN":    0,
		"DATA":       1,
		"CONNECT":    2,
		"DISCONNECT": 3,
		"ACK":        4,
	}
)

func (x ProxyHeader_Type) Enum() *ProxyHeader_Type {
	p := new(ProxyHeader_Type)
	*p = x
	return p
}

func (x ProxyHeader_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ProxyHeader_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_internal_proto_enumTypes[0].Descriptor()
}

func (ProxyHeader_Type) Type() protoreflect.EnumType {
	return &file_proto_internal_proto_enumTypes[0]
}

func (x ProxyHeader_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ProxyHeader_Type.Descriptor instead.
func (ProxyHeader_Type) EnumDescriptor() ([]byte, []int) {
	return file_proto_internal_proto_rawDescGZIP(), []int{0, 0}
}

type ProxyHeader struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ConnectionId int32            `protobuf:"varint,1,opt,name=connection_id,json=connectionId,proto3" json:"connection_id,omitempty"`
	Type         ProxyHeader_Type `protobuf:"varint,2,opt,name=type,proto3,enum=com.github.dlshle.dockman.ProxyHeader_Type" json:"type,omitempty"`
}

func (x *ProxyHeader) Reset() {
	*x = ProxyHeader{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_internal_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProxyHeader) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProxyHeader) ProtoMessage() {}

func (x *ProxyHeader) ProtoReflect() protoreflect.Message {
	mi := &file_proto_internal_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProxyHeader.ProtoReflect.Descriptor instead.
func (*ProxyHeader) Descriptor() ([]byte, []int) {
	return file_proto_internal_proto_rawDescGZIP(), []int{0}
}

func (x *ProxyHeader) GetConnectionId() int32 {
	if x != nil {
		return x.ConnectionId
	}
	return 0
}

func (x *ProxyHeader) GetType() ProxyHeader_Type {
	if x != nil {
		return x.Type
	}
	return ProxyHeader_UNKONWN
}

type ConnectRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Host string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	Port int32  `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
}

func (x *ConnectRequest) Reset() {
	*x = ConnectRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_internal_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConnectRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectRequest) ProtoMessage() {}

func (x *ConnectRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_internal_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectRequest.ProtoReflect.Descriptor instead.
func (*ConnectRequest) Descriptor() ([]byte, []int) {
	return file_proto_internal_proto_rawDescGZIP(), []int{1}
}

func (x *ConnectRequest) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *ConnectRequest) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

type DisconnectRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Reason string `protobuf:"bytes,1,opt,name=reason,proto3" json:"reason,omitempty"`
}

func (x *DisconnectRequest) Reset() {
	*x = DisconnectRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_internal_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DisconnectRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DisconnectRequest) ProtoMessage() {}

func (x *DisconnectRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_internal_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DisconnectRequest.ProtoReflect.Descriptor instead.
func (*DisconnectRequest) Descriptor() ([]byte, []int) {
	return file_proto_internal_proto_rawDescGZIP(), []int{2}
}

func (x *DisconnectRequest) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

type ProxyMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Header *ProxyHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	// Types that are assignable to Data:
	//	*ProxyMessage_Payload
	//	*ProxyMessage_ConnectRequest
	//	*ProxyMessage_DisconnectRequest
	Data isProxyMessage_Data `protobuf_oneof:"data"`
}

func (x *ProxyMessage) Reset() {
	*x = ProxyMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_internal_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProxyMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProxyMessage) ProtoMessage() {}

func (x *ProxyMessage) ProtoReflect() protoreflect.Message {
	mi := &file_proto_internal_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProxyMessage.ProtoReflect.Descriptor instead.
func (*ProxyMessage) Descriptor() ([]byte, []int) {
	return file_proto_internal_proto_rawDescGZIP(), []int{3}
}

func (x *ProxyMessage) GetHeader() *ProxyHeader {
	if x != nil {
		return x.Header
	}
	return nil
}

func (m *ProxyMessage) GetData() isProxyMessage_Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (x *ProxyMessage) GetPayload() []byte {
	if x, ok := x.GetData().(*ProxyMessage_Payload); ok {
		return x.Payload
	}
	return nil
}

func (x *ProxyMessage) GetConnectRequest() *ConnectRequest {
	if x, ok := x.GetData().(*ProxyMessage_ConnectRequest); ok {
		return x.ConnectRequest
	}
	return nil
}

func (x *ProxyMessage) GetDisconnectRequest() *DisconnectRequest {
	if x, ok := x.GetData().(*ProxyMessage_DisconnectRequest); ok {
		return x.DisconnectRequest
	}
	return nil
}

type isProxyMessage_Data interface {
	isProxyMessage_Data()
}

type ProxyMessage_Payload struct {
	Payload []byte `protobuf:"bytes,2,opt,name=payload,proto3,oneof"`
}

type ProxyMessage_ConnectRequest struct {
	ConnectRequest *ConnectRequest `protobuf:"bytes,3,opt,name=connect_request,json=connectRequest,proto3,oneof"`
}

type ProxyMessage_DisconnectRequest struct {
	DisconnectRequest *DisconnectRequest `protobuf:"bytes,4,opt,name=disconnect_request,json=disconnectRequest,proto3,oneof"`
}

func (*ProxyMessage_Payload) isProxyMessage_Data() {}

func (*ProxyMessage_ConnectRequest) isProxyMessage_Data() {}

func (*ProxyMessage_DisconnectRequest) isProxyMessage_Data() {}

var File_proto_internal_proto protoreflect.FileDescriptor

var file_proto_internal_proto_rawDesc = []byte{
	0x0a, 0x14, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x64, 0x6c, 0x73, 0x68, 0x6c, 0x65, 0x2e, 0x64, 0x6f, 0x63, 0x6b, 0x6d, 0x61,
	0x6e, 0x22, 0xb8, 0x01, 0x0a, 0x0b, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x12, 0x23, 0x0a, 0x0d, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x3f, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x2b, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x64, 0x6c, 0x73, 0x68, 0x6c, 0x65, 0x2e, 0x64, 0x6f, 0x63, 0x6b, 0x6d, 0x61, 0x6e,
	0x2e, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x54, 0x79, 0x70,
	0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x43, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4f, 0x4e, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04,
	0x44, 0x41, 0x54, 0x41, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x43, 0x4f, 0x4e, 0x4e, 0x45, 0x43,
	0x54, 0x10, 0x02, 0x12, 0x0e, 0x0a, 0x0a, 0x44, 0x49, 0x53, 0x43, 0x4f, 0x4e, 0x4e, 0x45, 0x43,
	0x54, 0x10, 0x03, 0x12, 0x07, 0x0a, 0x03, 0x41, 0x43, 0x4b, 0x10, 0x04, 0x22, 0x38, 0x0a, 0x0e,
	0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f,
	0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x22, 0x2b, 0x0a, 0x11, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x72,
	0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x61,
	0x73, 0x6f, 0x6e, 0x22, 0xa7, 0x02, 0x0a, 0x0c, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x3e, 0x0a, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x64, 0x6c, 0x73, 0x68, 0x6c, 0x65, 0x2e, 0x64, 0x6f, 0x63, 0x6b, 0x6d, 0x61, 0x6e,
	0x2e, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x06, 0x68, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64,
	0x12, 0x54, 0x0a, 0x0f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5f, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x63, 0x6f, 0x6d, 0x2e,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x64, 0x6c, 0x73, 0x68, 0x6c, 0x65, 0x2e, 0x64, 0x6f,
	0x63, 0x6b, 0x6d, 0x61, 0x6e, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x0e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x5d, 0x0a, 0x12, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x64, 0x6c, 0x73, 0x68, 0x6c, 0x65, 0x2e, 0x64, 0x6f, 0x63, 0x6b, 0x6d, 0x61, 0x6e, 0x2e, 0x44,
	0x69, 0x73, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x48, 0x00, 0x52, 0x11, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x42, 0x06, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x42, 0x21, 0x5a,
	0x1f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x6c, 0x73, 0x68,
	0x6c, 0x65, 0x2f, 0x64, 0x6f, 0x63, 0x6b, 0x6d, 0x61, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_internal_proto_rawDescOnce sync.Once
	file_proto_internal_proto_rawDescData = file_proto_internal_proto_rawDesc
)

func file_proto_internal_proto_rawDescGZIP() []byte {
	file_proto_internal_proto_rawDescOnce.Do(func() {
		file_proto_internal_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_internal_proto_rawDescData)
	})
	return file_proto_internal_proto_rawDescData
}

var file_proto_internal_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_internal_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proto_internal_proto_goTypes = []interface{}{
	(ProxyHeader_Type)(0),     // 0: com.github.dlshle.dockman.ProxyHeader.Type
	(*ProxyHeader)(nil),       // 1: com.github.dlshle.dockman.ProxyHeader
	(*ConnectRequest)(nil),    // 2: com.github.dlshle.dockman.ConnectRequest
	(*DisconnectRequest)(nil), // 3: com.github.dlshle.dockman.DisconnectRequest
	(*ProxyMessage)(nil),      // 4: com.github.dlshle.dockman.ProxyMessage
}
var file_proto_internal_proto_depIdxs = []int32{
	0, // 0: com.github.dlshle.dockman.ProxyHeader.type:type_name -> com.github.dlshle.dockman.ProxyHeader.Type
	1, // 1: com.github.dlshle.dockman.ProxyMessage.header:type_name -> com.github.dlshle.dockman.ProxyHeader
	2, // 2: com.github.dlshle.dockman.ProxyMessage.connect_request:type_name -> com.github.dlshle.dockman.ConnectRequest
	3, // 3: com.github.dlshle.dockman.ProxyMessage.disconnect_request:type_name -> com.github.dlshle.dockman.DisconnectRequest
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_proto_internal_proto_init() }
func file_proto_internal_proto_init() {
	if File_proto_internal_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_internal_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProxyHeader); i {
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
		file_proto_internal_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConnectRequest); i {
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
		file_proto_internal_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DisconnectRequest); i {
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
		file_proto_internal_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProxyMessage); i {
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
	file_proto_internal_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*ProxyMessage_Payload)(nil),
		(*ProxyMessage_ConnectRequest)(nil),
		(*ProxyMessage_DisconnectRequest)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_internal_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_internal_proto_goTypes,
		DependencyIndexes: file_proto_internal_proto_depIdxs,
		EnumInfos:         file_proto_internal_proto_enumTypes,
		MessageInfos:      file_proto_internal_proto_msgTypes,
	}.Build()
	File_proto_internal_proto = out.File
	file_proto_internal_proto_rawDesc = nil
	file_proto_internal_proto_goTypes = nil
	file_proto_internal_proto_depIdxs = nil
}
