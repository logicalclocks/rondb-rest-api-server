// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.2
// source: api/rdrs.proto

package api

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RonDBRESTClient is the client API for RonDBREST service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RonDBRESTClient interface {
	PKRead(ctx context.Context, in *PKReadRequestProto, opts ...grpc.CallOption) (*PKReadResponseProto, error)
	Batch(ctx context.Context, in *BatchRequestProto, opts ...grpc.CallOption) (*BatchResponseProto, error)
	Stat(ctx context.Context, in *StatRequestProto, opts ...grpc.CallOption) (*StatResponseProto, error)
}

type ronDBRESTClient struct {
	cc grpc.ClientConnInterface
}

func NewRonDBRESTClient(cc grpc.ClientConnInterface) RonDBRESTClient {
	return &ronDBRESTClient{cc}
}

func (c *ronDBRESTClient) PKRead(ctx context.Context, in *PKReadRequestProto, opts ...grpc.CallOption) (*PKReadResponseProto, error) {
	out := new(PKReadResponseProto)
	err := c.cc.Invoke(ctx, "/RonDBREST/PKRead", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ronDBRESTClient) Batch(ctx context.Context, in *BatchRequestProto, opts ...grpc.CallOption) (*BatchResponseProto, error) {
	out := new(BatchResponseProto)
	err := c.cc.Invoke(ctx, "/RonDBREST/Batch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ronDBRESTClient) Stat(ctx context.Context, in *StatRequestProto, opts ...grpc.CallOption) (*StatResponseProto, error) {
	out := new(StatResponseProto)
	err := c.cc.Invoke(ctx, "/RonDBREST/Stat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RonDBRESTServer is the server API for RonDBREST service.
// All implementations must embed UnimplementedRonDBRESTServer
// for forward compatibility
type RonDBRESTServer interface {
	PKRead(context.Context, *PKReadRequestProto) (*PKReadResponseProto, error)
	Batch(context.Context, *BatchRequestProto) (*BatchResponseProto, error)
	Stat(context.Context, *StatRequestProto) (*StatResponseProto, error)
	mustEmbedUnimplementedRonDBRESTServer()
}

// UnimplementedRonDBRESTServer must be embedded to have forward compatible implementations.
type UnimplementedRonDBRESTServer struct {
}

func (UnimplementedRonDBRESTServer) PKRead(context.Context, *PKReadRequestProto) (*PKReadResponseProto, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PKRead not implemented")
}
func (UnimplementedRonDBRESTServer) Batch(context.Context, *BatchRequestProto) (*BatchResponseProto, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Batch not implemented")
}
func (UnimplementedRonDBRESTServer) Stat(context.Context, *StatRequestProto) (*StatResponseProto, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stat not implemented")
}
func (UnimplementedRonDBRESTServer) mustEmbedUnimplementedRonDBRESTServer() {}

// UnsafeRonDBRESTServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RonDBRESTServer will
// result in compilation errors.
type UnsafeRonDBRESTServer interface {
	mustEmbedUnimplementedRonDBRESTServer()
}

func RegisterRonDBRESTServer(s grpc.ServiceRegistrar, srv RonDBRESTServer) {
	s.RegisterService(&RonDBREST_ServiceDesc, srv)
}

func _RonDBREST_PKRead_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PKReadRequestProto)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RonDBRESTServer).PKRead(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/RonDBREST/PKRead",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RonDBRESTServer).PKRead(ctx, req.(*PKReadRequestProto))
	}
	return interceptor(ctx, in, info, handler)
}

func _RonDBREST_Batch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchRequestProto)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RonDBRESTServer).Batch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/RonDBREST/Batch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RonDBRESTServer).Batch(ctx, req.(*BatchRequestProto))
	}
	return interceptor(ctx, in, info, handler)
}

func _RonDBREST_Stat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatRequestProto)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RonDBRESTServer).Stat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/RonDBREST/Stat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RonDBRESTServer).Stat(ctx, req.(*StatRequestProto))
	}
	return interceptor(ctx, in, info, handler)
}

// RonDBREST_ServiceDesc is the grpc.ServiceDesc for RonDBREST service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RonDBREST_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "RonDBREST",
	HandlerType: (*RonDBRESTServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PKRead",
			Handler:    _RonDBREST_PKRead_Handler,
		},
		{
			MethodName: "Batch",
			Handler:    _RonDBREST_Batch_Handler,
		},
		{
			MethodName: "Stat",
			Handler:    _RonDBREST_Stat_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/rdrs.proto",
}
