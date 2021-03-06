//
// This file is part of the RonDB REST API Server
// Copyright (c) 2022 Hopsworks AB
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.2
// source: api/rdrs.proto

package api

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

//__________________  PK Read Operation __________________
type FilterProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Column *string `protobuf:"bytes,1,req,name=Column" json:"Column,omitempty"`
	Value  *string `protobuf:"bytes,2,req,name=Value" json:"Value,omitempty"`
}

func (x *FilterProto) Reset() {
	*x = FilterProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_rdrs_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FilterProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FilterProto) ProtoMessage() {}

func (x *FilterProto) ProtoReflect() protoreflect.Message {
	mi := &file_api_rdrs_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FilterProto.ProtoReflect.Descriptor instead.
func (*FilterProto) Descriptor() ([]byte, []int) {
	return file_api_rdrs_proto_rawDescGZIP(), []int{0}
}

func (x *FilterProto) GetColumn() string {
	if x != nil && x.Column != nil {
		return *x.Column
	}
	return ""
}

func (x *FilterProto) GetValue() string {
	if x != nil && x.Value != nil {
		return *x.Value
	}
	return ""
}

type ReadColumnProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Column         *string `protobuf:"bytes,1,req,name=Column" json:"Column,omitempty"`
	DataReturnType *string `protobuf:"bytes,2,opt,name=DataReturnType" json:"DataReturnType,omitempty"`
}

func (x *ReadColumnProto) Reset() {
	*x = ReadColumnProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_rdrs_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadColumnProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadColumnProto) ProtoMessage() {}

func (x *ReadColumnProto) ProtoReflect() protoreflect.Message {
	mi := &file_api_rdrs_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadColumnProto.ProtoReflect.Descriptor instead.
func (*ReadColumnProto) Descriptor() ([]byte, []int) {
	return file_api_rdrs_proto_rawDescGZIP(), []int{1}
}

func (x *ReadColumnProto) GetColumn() string {
	if x != nil && x.Column != nil {
		return *x.Column
	}
	return ""
}

func (x *ReadColumnProto) GetDataReturnType() string {
	if x != nil && x.DataReturnType != nil {
		return *x.DataReturnType
	}
	return ""
}

type PKReadRequestProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	APIKey      *string            `protobuf:"bytes,1,opt,name=APIKey" json:"APIKey,omitempty"`
	DB          *string            `protobuf:"bytes,2,req,name=DB" json:"DB,omitempty"`
	Table       *string            `protobuf:"bytes,3,req,name=Table" json:"Table,omitempty"`
	Filters     []*FilterProto     `protobuf:"bytes,4,rep,name=Filters" json:"Filters,omitempty"`
	ReadColumns []*ReadColumnProto `protobuf:"bytes,5,rep,name=ReadColumns" json:"ReadColumns,omitempty"`
	OperationID *string            `protobuf:"bytes,6,opt,name=OperationID" json:"OperationID,omitempty"`
}

func (x *PKReadRequestProto) Reset() {
	*x = PKReadRequestProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_rdrs_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PKReadRequestProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PKReadRequestProto) ProtoMessage() {}

func (x *PKReadRequestProto) ProtoReflect() protoreflect.Message {
	mi := &file_api_rdrs_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PKReadRequestProto.ProtoReflect.Descriptor instead.
func (*PKReadRequestProto) Descriptor() ([]byte, []int) {
	return file_api_rdrs_proto_rawDescGZIP(), []int{2}
}

func (x *PKReadRequestProto) GetAPIKey() string {
	if x != nil && x.APIKey != nil {
		return *x.APIKey
	}
	return ""
}

func (x *PKReadRequestProto) GetDB() string {
	if x != nil && x.DB != nil {
		return *x.DB
	}
	return ""
}

func (x *PKReadRequestProto) GetTable() string {
	if x != nil && x.Table != nil {
		return *x.Table
	}
	return ""
}

func (x *PKReadRequestProto) GetFilters() []*FilterProto {
	if x != nil {
		return x.Filters
	}
	return nil
}

func (x *PKReadRequestProto) GetReadColumns() []*ReadColumnProto {
	if x != nil {
		return x.ReadColumns
	}
	return nil
}

func (x *PKReadRequestProto) GetOperationID() string {
	if x != nil && x.OperationID != nil {
		return *x.OperationID
	}
	return ""
}

type ColumnValueProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name *string `protobuf:"bytes,1,opt,name=Name" json:"Name,omitempty"`
}

func (x *ColumnValueProto) Reset() {
	*x = ColumnValueProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_rdrs_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ColumnValueProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ColumnValueProto) ProtoMessage() {}

func (x *ColumnValueProto) ProtoReflect() protoreflect.Message {
	mi := &file_api_rdrs_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ColumnValueProto.ProtoReflect.Descriptor instead.
func (*ColumnValueProto) Descriptor() ([]byte, []int) {
	return file_api_rdrs_proto_rawDescGZIP(), []int{3}
}

func (x *ColumnValueProto) GetName() string {
	if x != nil && x.Name != nil {
		return *x.Name
	}
	return ""
}

type PKReadResponseProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OperationID *string                      `protobuf:"bytes,1,opt,name=OperationID" json:"OperationID,omitempty"`
	Code        *int32                       `protobuf:"varint,2,opt,name=code" json:"code,omitempty"`
	Data        map[string]*ColumnValueProto `protobuf:"bytes,3,rep,name=Data" json:"Data,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (x *PKReadResponseProto) Reset() {
	*x = PKReadResponseProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_rdrs_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PKReadResponseProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PKReadResponseProto) ProtoMessage() {}

func (x *PKReadResponseProto) ProtoReflect() protoreflect.Message {
	mi := &file_api_rdrs_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PKReadResponseProto.ProtoReflect.Descriptor instead.
func (*PKReadResponseProto) Descriptor() ([]byte, []int) {
	return file_api_rdrs_proto_rawDescGZIP(), []int{4}
}

func (x *PKReadResponseProto) GetOperationID() string {
	if x != nil && x.OperationID != nil {
		return *x.OperationID
	}
	return ""
}

func (x *PKReadResponseProto) GetCode() int32 {
	if x != nil && x.Code != nil {
		return *x.Code
	}
	return 0
}

func (x *PKReadResponseProto) GetData() map[string]*ColumnValueProto {
	if x != nil {
		return x.Data
	}
	return nil
}

//__________________  Batch Operation ________________________
type BatchRequestProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	APIKey     *string               `protobuf:"bytes,1,opt,name=APIKey" json:"APIKey,omitempty"`
	Operations []*PKReadRequestProto `protobuf:"bytes,2,rep,name=operations" json:"operations,omitempty"`
}

func (x *BatchRequestProto) Reset() {
	*x = BatchRequestProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_rdrs_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BatchRequestProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchRequestProto) ProtoMessage() {}

func (x *BatchRequestProto) ProtoReflect() protoreflect.Message {
	mi := &file_api_rdrs_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchRequestProto.ProtoReflect.Descriptor instead.
func (*BatchRequestProto) Descriptor() ([]byte, []int) {
	return file_api_rdrs_proto_rawDescGZIP(), []int{5}
}

func (x *BatchRequestProto) GetAPIKey() string {
	if x != nil && x.APIKey != nil {
		return *x.APIKey
	}
	return ""
}

func (x *BatchRequestProto) GetOperations() []*PKReadRequestProto {
	if x != nil {
		return x.Operations
	}
	return nil
}

type BatchResponseProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Responses []*PKReadResponseProto `protobuf:"bytes,1,rep,name=responses" json:"responses,omitempty"`
}

func (x *BatchResponseProto) Reset() {
	*x = BatchResponseProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_rdrs_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BatchResponseProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchResponseProto) ProtoMessage() {}

func (x *BatchResponseProto) ProtoReflect() protoreflect.Message {
	mi := &file_api_rdrs_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchResponseProto.ProtoReflect.Descriptor instead.
func (*BatchResponseProto) Descriptor() ([]byte, []int) {
	return file_api_rdrs_proto_rawDescGZIP(), []int{6}
}

func (x *BatchResponseProto) GetResponses() []*PKReadResponseProto {
	if x != nil {
		return x.Responses
	}
	return nil
}

type MemoryStatsProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AllocationsCount   *int64 `protobuf:"varint,1,req,name=AllocationsCount" json:"AllocationsCount,omitempty"`
	DeallocationsCount *int64 `protobuf:"varint,2,req,name=DeallocationsCount" json:"DeallocationsCount,omitempty"`
	BuffersCount       *int64 `protobuf:"varint,3,req,name=BuffersCount" json:"BuffersCount,omitempty"`
	FreeBuffers        *int64 `protobuf:"varint,4,req,name=FreeBuffers" json:"FreeBuffers,omitempty"`
}

func (x *MemoryStatsProto) Reset() {
	*x = MemoryStatsProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_rdrs_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MemoryStatsProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MemoryStatsProto) ProtoMessage() {}

func (x *MemoryStatsProto) ProtoReflect() protoreflect.Message {
	mi := &file_api_rdrs_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MemoryStatsProto.ProtoReflect.Descriptor instead.
func (*MemoryStatsProto) Descriptor() ([]byte, []int) {
	return file_api_rdrs_proto_rawDescGZIP(), []int{7}
}

func (x *MemoryStatsProto) GetAllocationsCount() int64 {
	if x != nil && x.AllocationsCount != nil {
		return *x.AllocationsCount
	}
	return 0
}

func (x *MemoryStatsProto) GetDeallocationsCount() int64 {
	if x != nil && x.DeallocationsCount != nil {
		return *x.DeallocationsCount
	}
	return 0
}

func (x *MemoryStatsProto) GetBuffersCount() int64 {
	if x != nil && x.BuffersCount != nil {
		return *x.BuffersCount
	}
	return 0
}

func (x *MemoryStatsProto) GetFreeBuffers() int64 {
	if x != nil && x.FreeBuffers != nil {
		return *x.FreeBuffers
	}
	return 0
}

type RonDBStatsProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NdbObjectsCreationCount *int64 `protobuf:"varint,1,req,name=NdbObjectsCreationCount" json:"NdbObjectsCreationCount,omitempty"`
	NdbObjectsDeletionCount *int64 `protobuf:"varint,2,req,name=NdbObjectsDeletionCount" json:"NdbObjectsDeletionCount,omitempty"`
	NdbObjectsTotalCount    *int64 `protobuf:"varint,3,req,name=NdbObjectsTotalCount" json:"NdbObjectsTotalCount,omitempty"`
	NdbObjectsFreeCount     *int64 `protobuf:"varint,4,req,name=NdbObjectsFreeCount" json:"NdbObjectsFreeCount,omitempty"`
}

func (x *RonDBStatsProto) Reset() {
	*x = RonDBStatsProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_rdrs_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RonDBStatsProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RonDBStatsProto) ProtoMessage() {}

func (x *RonDBStatsProto) ProtoReflect() protoreflect.Message {
	mi := &file_api_rdrs_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RonDBStatsProto.ProtoReflect.Descriptor instead.
func (*RonDBStatsProto) Descriptor() ([]byte, []int) {
	return file_api_rdrs_proto_rawDescGZIP(), []int{8}
}

func (x *RonDBStatsProto) GetNdbObjectsCreationCount() int64 {
	if x != nil && x.NdbObjectsCreationCount != nil {
		return *x.NdbObjectsCreationCount
	}
	return 0
}

func (x *RonDBStatsProto) GetNdbObjectsDeletionCount() int64 {
	if x != nil && x.NdbObjectsDeletionCount != nil {
		return *x.NdbObjectsDeletionCount
	}
	return 0
}

func (x *RonDBStatsProto) GetNdbObjectsTotalCount() int64 {
	if x != nil && x.NdbObjectsTotalCount != nil {
		return *x.NdbObjectsTotalCount
	}
	return 0
}

func (x *RonDBStatsProto) GetNdbObjectsFreeCount() int64 {
	if x != nil && x.NdbObjectsFreeCount != nil {
		return *x.NdbObjectsFreeCount
	}
	return 0
}

type StatRequestProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *StatRequestProto) Reset() {
	*x = StatRequestProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_rdrs_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatRequestProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatRequestProto) ProtoMessage() {}

func (x *StatRequestProto) ProtoReflect() protoreflect.Message {
	mi := &file_api_rdrs_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatRequestProto.ProtoReflect.Descriptor instead.
func (*StatRequestProto) Descriptor() ([]byte, []int) {
	return file_api_rdrs_proto_rawDescGZIP(), []int{9}
}

type StatResponseProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MemoryStats *MemoryStatsProto `protobuf:"bytes,1,req,name=MemoryStats" json:"MemoryStats,omitempty"`
	RonDBStats  *RonDBStatsProto  `protobuf:"bytes,2,req,name=RonDBStats" json:"RonDBStats,omitempty"`
}

func (x *StatResponseProto) Reset() {
	*x = StatResponseProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_rdrs_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatResponseProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatResponseProto) ProtoMessage() {}

func (x *StatResponseProto) ProtoReflect() protoreflect.Message {
	mi := &file_api_rdrs_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatResponseProto.ProtoReflect.Descriptor instead.
func (*StatResponseProto) Descriptor() ([]byte, []int) {
	return file_api_rdrs_proto_rawDescGZIP(), []int{10}
}

func (x *StatResponseProto) GetMemoryStats() *MemoryStatsProto {
	if x != nil {
		return x.MemoryStats
	}
	return nil
}

func (x *StatResponseProto) GetRonDBStats() *RonDBStatsProto {
	if x != nil {
		return x.RonDBStats
	}
	return nil
}

var File_api_rdrs_proto protoreflect.FileDescriptor

var file_api_rdrs_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x64, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x3b, 0x0a, 0x0b, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x16, 0x0a, 0x06, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x18, 0x01, 0x20, 0x02, 0x28, 0x09, 0x52,
	0x06, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x02, 0x28, 0x09, 0x52, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x51, 0x0a,
	0x0f, 0x52, 0x65, 0x61, 0x64, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x16, 0x0a, 0x06, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x18, 0x01, 0x20, 0x02, 0x28, 0x09,
	0x52, 0x06, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x12, 0x26, 0x0a, 0x0e, 0x44, 0x61, 0x74, 0x61,
	0x52, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0e, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x54, 0x79, 0x70, 0x65,
	0x22, 0xd0, 0x01, 0x0a, 0x12, 0x50, 0x4b, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x0a, 0x06, 0x41, 0x50, 0x49, 0x4b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x41, 0x50, 0x49, 0x4b, 0x65, 0x79, 0x12,
	0x0e, 0x0a, 0x02, 0x44, 0x42, 0x18, 0x02, 0x20, 0x02, 0x28, 0x09, 0x52, 0x02, 0x44, 0x42, 0x12,
	0x14, 0x0a, 0x05, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x02, 0x28, 0x09, 0x52, 0x05,
	0x54, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x26, 0x0a, 0x07, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73,
	0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x52, 0x07, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x12, 0x32, 0x0a,
	0x0b, 0x52, 0x65, 0x61, 0x64, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x73, 0x18, 0x05, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x10, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x52, 0x0b, 0x52, 0x65, 0x61, 0x64, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e,
	0x73, 0x12, 0x20, 0x0a, 0x0b, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x44,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x49, 0x44, 0x22, 0x26, 0x0a, 0x10, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0xcb, 0x01, 0x0a, 0x13,
	0x50, 0x4b, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x20, 0x0a, 0x0b, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x32, 0x0a, 0x04, 0x44, 0x61, 0x74,
	0x61, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x50, 0x4b, 0x52, 0x65, 0x61, 0x64,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x44, 0x61,
	0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x4a, 0x0a,
	0x09, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x27, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x43, 0x6f,
	0x6c, 0x75, 0x6d, 0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x60, 0x0a, 0x11, 0x42, 0x61, 0x74,
	0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16,
	0x0a, 0x06, 0x41, 0x50, 0x49, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x41, 0x50, 0x49, 0x4b, 0x65, 0x79, 0x12, 0x33, 0x0a, 0x0a, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x50, 0x4b, 0x52,
	0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52,
	0x0a, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x48, 0x0a, 0x12, 0x42,
	0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x32, 0x0a, 0x09, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x50, 0x4b, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x09, 0x72, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x73, 0x22, 0xb4, 0x01, 0x0a, 0x10, 0x4d, 0x65, 0x6d, 0x6f, 0x72, 0x79,
	0x53, 0x74, 0x61, 0x74, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x2a, 0x0a, 0x10, 0x41, 0x6c,
	0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01,
	0x20, 0x02, 0x28, 0x03, 0x52, 0x10, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x2e, 0x0a, 0x12, 0x44, 0x65, 0x61, 0x6c, 0x6c, 0x6f,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x02,
	0x28, 0x03, 0x52, 0x12, 0x44, 0x65, 0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x22, 0x0a, 0x0c, 0x42, 0x75, 0x66, 0x66, 0x65, 0x72,
	0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x02, 0x28, 0x03, 0x52, 0x0c, 0x42, 0x75,
	0x66, 0x66, 0x65, 0x72, 0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x46, 0x72,
	0x65, 0x65, 0x42, 0x75, 0x66, 0x66, 0x65, 0x72, 0x73, 0x18, 0x04, 0x20, 0x02, 0x28, 0x03, 0x52,
	0x0b, 0x46, 0x72, 0x65, 0x65, 0x42, 0x75, 0x66, 0x66, 0x65, 0x72, 0x73, 0x22, 0xeb, 0x01, 0x0a,
	0x0f, 0x52, 0x6f, 0x6e, 0x44, 0x42, 0x53, 0x74, 0x61, 0x74, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x38, 0x0a, 0x17, 0x4e, 0x64, 0x62, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x02, 0x28,
	0x03, 0x52, 0x17, 0x4e, 0x64, 0x62, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x38, 0x0a, 0x17, 0x4e, 0x64,
	0x62, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e,
	0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x02, 0x28, 0x03, 0x52, 0x17, 0x4e, 0x64, 0x62,
	0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x43,
	0x6f, 0x75, 0x6e, 0x74, 0x12, 0x32, 0x0a, 0x14, 0x4e, 0x64, 0x62, 0x4f, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x73, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x02,
	0x28, 0x03, 0x52, 0x14, 0x4e, 0x64, 0x62, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x54, 0x6f,
	0x74, 0x61, 0x6c, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x30, 0x0a, 0x13, 0x4e, 0x64, 0x62, 0x4f,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x46, 0x72, 0x65, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18,
	0x04, 0x20, 0x02, 0x28, 0x03, 0x52, 0x13, 0x4e, 0x64, 0x62, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74,
	0x73, 0x46, 0x72, 0x65, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x12, 0x0a, 0x10, 0x53, 0x74,
	0x61, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x7a,
	0x0a, 0x11, 0x53, 0x74, 0x61, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x33, 0x0a, 0x0b, 0x4d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x53, 0x74, 0x61,
	0x74, 0x73, 0x18, 0x01, 0x20, 0x02, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x4d, 0x65, 0x6d, 0x6f, 0x72,
	0x79, 0x53, 0x74, 0x61, 0x74, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x0b, 0x4d, 0x65, 0x6d,
	0x6f, 0x72, 0x79, 0x53, 0x74, 0x61, 0x74, 0x73, 0x12, 0x30, 0x0a, 0x0a, 0x52, 0x6f, 0x6e, 0x44,
	0x42, 0x53, 0x74, 0x61, 0x74, 0x73, 0x18, 0x02, 0x20, 0x02, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x52,
	0x6f, 0x6e, 0x44, 0x42, 0x53, 0x74, 0x61, 0x74, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x0a,
	0x52, 0x6f, 0x6e, 0x44, 0x42, 0x53, 0x74, 0x61, 0x74, 0x73, 0x32, 0xa1, 0x01, 0x0a, 0x09, 0x52,
	0x6f, 0x6e, 0x44, 0x42, 0x52, 0x45, 0x53, 0x54, 0x12, 0x33, 0x0a, 0x06, 0x50, 0x4b, 0x52, 0x65,
	0x61, 0x64, 0x12, 0x13, 0x2e, 0x50, 0x4b, 0x52, 0x65, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x14, 0x2e, 0x50, 0x4b, 0x52, 0x65, 0x61, 0x64,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x30, 0x0a,
	0x05, 0x42, 0x61, 0x74, 0x63, 0x68, 0x12, 0x12, 0x2e, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x13, 0x2e, 0x42, 0x61, 0x74,
	0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x2d, 0x0a, 0x04, 0x53, 0x74, 0x61, 0x74, 0x12, 0x11, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x12, 0x2e, 0x53, 0x74, 0x61,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x42, 0x0b,
	0x5a, 0x09, 0x2e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70, 0x69,
}

var (
	file_api_rdrs_proto_rawDescOnce sync.Once
	file_api_rdrs_proto_rawDescData = file_api_rdrs_proto_rawDesc
)

func file_api_rdrs_proto_rawDescGZIP() []byte {
	file_api_rdrs_proto_rawDescOnce.Do(func() {
		file_api_rdrs_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_rdrs_proto_rawDescData)
	})
	return file_api_rdrs_proto_rawDescData
}

var file_api_rdrs_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_api_rdrs_proto_goTypes = []interface{}{
	(*FilterProto)(nil),         // 0: FilterProto
	(*ReadColumnProto)(nil),     // 1: ReadColumnProto
	(*PKReadRequestProto)(nil),  // 2: PKReadRequestProto
	(*ColumnValueProto)(nil),    // 3: ColumnValueProto
	(*PKReadResponseProto)(nil), // 4: PKReadResponseProto
	(*BatchRequestProto)(nil),   // 5: BatchRequestProto
	(*BatchResponseProto)(nil),  // 6: BatchResponseProto
	(*MemoryStatsProto)(nil),    // 7: MemoryStatsProto
	(*RonDBStatsProto)(nil),     // 8: RonDBStatsProto
	(*StatRequestProto)(nil),    // 9: StatRequestProto
	(*StatResponseProto)(nil),   // 10: StatResponseProto
	nil,                         // 11: PKReadResponseProto.DataEntry
}
var file_api_rdrs_proto_depIdxs = []int32{
	0,  // 0: PKReadRequestProto.Filters:type_name -> FilterProto
	1,  // 1: PKReadRequestProto.ReadColumns:type_name -> ReadColumnProto
	11, // 2: PKReadResponseProto.Data:type_name -> PKReadResponseProto.DataEntry
	2,  // 3: BatchRequestProto.operations:type_name -> PKReadRequestProto
	4,  // 4: BatchResponseProto.responses:type_name -> PKReadResponseProto
	7,  // 5: StatResponseProto.MemoryStats:type_name -> MemoryStatsProto
	8,  // 6: StatResponseProto.RonDBStats:type_name -> RonDBStatsProto
	3,  // 7: PKReadResponseProto.DataEntry.value:type_name -> ColumnValueProto
	2,  // 8: RonDBREST.PKRead:input_type -> PKReadRequestProto
	5,  // 9: RonDBREST.Batch:input_type -> BatchRequestProto
	9,  // 10: RonDBREST.Stat:input_type -> StatRequestProto
	4,  // 11: RonDBREST.PKRead:output_type -> PKReadResponseProto
	6,  // 12: RonDBREST.Batch:output_type -> BatchResponseProto
	10, // 13: RonDBREST.Stat:output_type -> StatResponseProto
	11, // [11:14] is the sub-list for method output_type
	8,  // [8:11] is the sub-list for method input_type
	8,  // [8:8] is the sub-list for extension type_name
	8,  // [8:8] is the sub-list for extension extendee
	0,  // [0:8] is the sub-list for field type_name
}

func init() { file_api_rdrs_proto_init() }
func file_api_rdrs_proto_init() {
	if File_api_rdrs_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_rdrs_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FilterProto); i {
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
		file_api_rdrs_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadColumnProto); i {
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
		file_api_rdrs_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PKReadRequestProto); i {
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
		file_api_rdrs_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ColumnValueProto); i {
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
		file_api_rdrs_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PKReadResponseProto); i {
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
		file_api_rdrs_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BatchRequestProto); i {
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
		file_api_rdrs_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BatchResponseProto); i {
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
		file_api_rdrs_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MemoryStatsProto); i {
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
		file_api_rdrs_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RonDBStatsProto); i {
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
		file_api_rdrs_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatRequestProto); i {
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
		file_api_rdrs_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatResponseProto); i {
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
			RawDescriptor: file_api_rdrs_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_rdrs_proto_goTypes,
		DependencyIndexes: file_api_rdrs_proto_depIdxs,
		MessageInfos:      file_api_rdrs_proto_msgTypes,
	}.Build()
	File_api_rdrs_proto = out.File
	file_api_rdrs_proto_rawDesc = nil
	file_api_rdrs_proto_goTypes = nil
	file_api_rdrs_proto_depIdxs = nil
}
