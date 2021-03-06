// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

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

// RequestClient is the client API for Request service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RequestClient interface {
	List(ctx context.Context, in *Rule, opts ...grpc.CallOption) (*ListResp, error)
	Update(ctx context.Context, in *Rule, opts ...grpc.CallOption) (*ErrResp, error)
	Remove(ctx context.Context, in *Rule, opts ...grpc.CallOption) (*ErrResp, error)
}

type requestClient struct {
	cc grpc.ClientConnInterface
}

func NewRequestClient(cc grpc.ClientConnInterface) RequestClient {
	return &requestClient{cc}
}

func (c *requestClient) List(ctx context.Context, in *Rule, opts ...grpc.CallOption) (*ListResp, error) {
	out := new(ListResp)
	err := c.cc.Invoke(ctx, "/proto.Request/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *requestClient) Update(ctx context.Context, in *Rule, opts ...grpc.CallOption) (*ErrResp, error) {
	out := new(ErrResp)
	err := c.cc.Invoke(ctx, "/proto.Request/Update", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *requestClient) Remove(ctx context.Context, in *Rule, opts ...grpc.CallOption) (*ErrResp, error) {
	out := new(ErrResp)
	err := c.cc.Invoke(ctx, "/proto.Request/Remove", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RequestServer is the server API for Request service.
// All implementations must embed UnimplementedRequestServer
// for forward compatibility
type RequestServer interface {
	List(context.Context, *Rule) (*ListResp, error)
	Update(context.Context, *Rule) (*ErrResp, error)
	Remove(context.Context, *Rule) (*ErrResp, error)
	mustEmbedUnimplementedRequestServer()
}

// UnimplementedRequestServer must be embedded to have forward compatible implementations.
type UnimplementedRequestServer struct {
}

func (UnimplementedRequestServer) List(context.Context, *Rule) (*ListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedRequestServer) Update(context.Context, *Rule) (*ErrResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}
func (UnimplementedRequestServer) Remove(context.Context, *Rule) (*ErrResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Remove not implemented")
}
func (UnimplementedRequestServer) mustEmbedUnimplementedRequestServer() {}

// UnsafeRequestServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RequestServer will
// result in compilation errors.
type UnsafeRequestServer interface {
	mustEmbedUnimplementedRequestServer()
}

func RegisterRequestServer(s grpc.ServiceRegistrar, srv RequestServer) {
	s.RegisterService(&Request_ServiceDesc, srv)
}

func _Request_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Rule)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RequestServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Request/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RequestServer).List(ctx, req.(*Rule))
	}
	return interceptor(ctx, in, info, handler)
}

func _Request_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Rule)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RequestServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Request/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RequestServer).Update(ctx, req.(*Rule))
	}
	return interceptor(ctx, in, info, handler)
}

func _Request_Remove_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Rule)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RequestServer).Remove(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Request/Remove",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RequestServer).Remove(ctx, req.(*Rule))
	}
	return interceptor(ctx, in, info, handler)
}

// Request_ServiceDesc is the grpc.ServiceDesc for Request service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Request_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Request",
	HandlerType: (*RequestServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _Request_List_Handler,
		},
		{
			MethodName: "Update",
			Handler:    _Request_Update_Handler,
		},
		{
			MethodName: "Remove",
			Handler:    _Request_Remove_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc.proto",
}
