// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v5.29.3
// source: resources/protobuf/service.proto

package grpctest

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetItemRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int32                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetItemRequest) Reset() {
	*x = GetItemRequest{}
	mi := &file_resources_protobuf_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetItemRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetItemRequest) ProtoMessage() {}

func (x *GetItemRequest) ProtoReflect() protoreflect.Message {
	mi := &file_resources_protobuf_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetItemRequest.ProtoReflect.Descriptor instead.
func (*GetItemRequest) Descriptor() ([]byte, []int) {
	return file_resources_protobuf_service_proto_rawDescGZIP(), []int{0}
}

func (x *GetItemRequest) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type ListItemsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	PageSize      int32                  `protobuf:"varint,1,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListItemsRequest) Reset() {
	*x = ListItemsRequest{}
	mi := &file_resources_protobuf_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListItemsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListItemsRequest) ProtoMessage() {}

func (x *ListItemsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_resources_protobuf_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListItemsRequest.ProtoReflect.Descriptor instead.
func (*ListItemsRequest) Descriptor() ([]byte, []int) {
	return file_resources_protobuf_service_proto_rawDescGZIP(), []int{1}
}

func (x *ListItemsRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

type CreateItemsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	NumItems      int64                  `protobuf:"varint,1,opt,name=num_items,json=numItems,proto3" json:"num_items,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateItemsResponse) Reset() {
	*x = CreateItemsResponse{}
	mi := &file_resources_protobuf_service_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateItemsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateItemsResponse) ProtoMessage() {}

func (x *CreateItemsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_resources_protobuf_service_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateItemsResponse.ProtoReflect.Descriptor instead.
func (*CreateItemsResponse) Descriptor() ([]byte, []int) {
	return file_resources_protobuf_service_proto_rawDescGZIP(), []int{2}
}

func (x *CreateItemsResponse) GetNumItems() int64 {
	if x != nil {
		return x.NumItems
	}
	return 0
}

type Item struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int32                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Locale        string                 `protobuf:"bytes,2,opt,name=locale,proto3" json:"locale,omitempty"`
	Name          string                 `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	CreateTime    *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Item) Reset() {
	*x = Item{}
	mi := &file_resources_protobuf_service_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Item) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Item) ProtoMessage() {}

func (x *Item) ProtoReflect() protoreflect.Message {
	mi := &file_resources_protobuf_service_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Item.ProtoReflect.Descriptor instead.
func (*Item) Descriptor() ([]byte, []int) {
	return file_resources_protobuf_service_proto_rawDescGZIP(), []int{3}
}

func (x *Item) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Item) GetLocale() string {
	if x != nil {
		return x.Locale
	}
	return ""
}

func (x *Item) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Item) GetCreateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CreateTime
	}
	return nil
}

var File_resources_protobuf_service_proto protoreflect.FileDescriptor

var file_resources_protobuf_service_proto_rawDesc = string([]byte{
	0x0a, 0x20, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x08, 0x67, 0x72, 0x70, 0x63, 0x74, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x20, 0x0a,
	0x0e, 0x47, 0x65, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x69, 0x64, 0x22,
	0x2f, 0x0a, 0x10, 0x4c, 0x69, 0x73, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65,
	0x22, 0x32, 0x0a, 0x13, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x6e, 0x75, 0x6d, 0x5f, 0x69,
	0x74, 0x65, 0x6d, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x6e, 0x75, 0x6d, 0x49,
	0x74, 0x65, 0x6d, 0x73, 0x22, 0x7f, 0x0a, 0x04, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06,
	0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6c, 0x6f,
	0x63, 0x61, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x3b, 0x0a, 0x0b, 0x63, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x54, 0x69, 0x6d, 0x65, 0x32, 0xf3, 0x01, 0x0a, 0x0b, 0x49, 0x74, 0x65, 0x6d, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x33, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x49, 0x74, 0x65, 0x6d,
	0x12, 0x18, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x47, 0x65, 0x74, 0x49,
	0x74, 0x65, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x67, 0x72, 0x70,
	0x63, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x39, 0x0a, 0x09, 0x4c, 0x69,
	0x73, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x12, 0x1a, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x74, 0x65,
	0x73, 0x74, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x49,
	0x74, 0x65, 0x6d, 0x30, 0x01, 0x12, 0x3e, 0x0a, 0x0b, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x49,
	0x74, 0x65, 0x6d, 0x73, 0x12, 0x0e, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x74, 0x65, 0x73, 0x74, 0x2e,
	0x49, 0x74, 0x65, 0x6d, 0x1a, 0x1d, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x74, 0x65, 0x73, 0x74, 0x2e,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x28, 0x01, 0x12, 0x34, 0x0a, 0x0e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f,
	0x72, 0x6d, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x12, 0x0e, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x74, 0x65,
	0x73, 0x74, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x1a, 0x0e, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x74, 0x65,
	0x73, 0x74, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x28, 0x01, 0x30, 0x01, 0x42, 0x19, 0x5a, 0x17, 0x74,
	0x65, 0x73, 0x74, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x3b, 0x67, 0x72,
	0x70, 0x63, 0x74, 0x65, 0x73, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_resources_protobuf_service_proto_rawDescOnce sync.Once
	file_resources_protobuf_service_proto_rawDescData []byte
)

func file_resources_protobuf_service_proto_rawDescGZIP() []byte {
	file_resources_protobuf_service_proto_rawDescOnce.Do(func() {
		file_resources_protobuf_service_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_resources_protobuf_service_proto_rawDesc), len(file_resources_protobuf_service_proto_rawDesc)))
	})
	return file_resources_protobuf_service_proto_rawDescData
}

var file_resources_protobuf_service_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_resources_protobuf_service_proto_goTypes = []any{
	(*GetItemRequest)(nil),        // 0: grpctest.GetItemRequest
	(*ListItemsRequest)(nil),      // 1: grpctest.ListItemsRequest
	(*CreateItemsResponse)(nil),   // 2: grpctest.CreateItemsResponse
	(*Item)(nil),                  // 3: grpctest.Item
	(*timestamppb.Timestamp)(nil), // 4: google.protobuf.Timestamp
}
var file_resources_protobuf_service_proto_depIdxs = []int32{
	4, // 0: grpctest.Item.create_time:type_name -> google.protobuf.Timestamp
	0, // 1: grpctest.ItemService.GetItem:input_type -> grpctest.GetItemRequest
	1, // 2: grpctest.ItemService.ListItems:input_type -> grpctest.ListItemsRequest
	3, // 3: grpctest.ItemService.CreateItems:input_type -> grpctest.Item
	3, // 4: grpctest.ItemService.TransformItems:input_type -> grpctest.Item
	3, // 5: grpctest.ItemService.GetItem:output_type -> grpctest.Item
	3, // 6: grpctest.ItemService.ListItems:output_type -> grpctest.Item
	2, // 7: grpctest.ItemService.CreateItems:output_type -> grpctest.CreateItemsResponse
	3, // 8: grpctest.ItemService.TransformItems:output_type -> grpctest.Item
	5, // [5:9] is the sub-list for method output_type
	1, // [1:5] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_resources_protobuf_service_proto_init() }
func file_resources_protobuf_service_proto_init() {
	if File_resources_protobuf_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_resources_protobuf_service_proto_rawDesc), len(file_resources_protobuf_service_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_resources_protobuf_service_proto_goTypes,
		DependencyIndexes: file_resources_protobuf_service_proto_depIdxs,
		MessageInfos:      file_resources_protobuf_service_proto_msgTypes,
	}.Build()
	File_resources_protobuf_service_proto = out.File
	file_resources_protobuf_service_proto_goTypes = nil
	file_resources_protobuf_service_proto_depIdxs = nil
}
