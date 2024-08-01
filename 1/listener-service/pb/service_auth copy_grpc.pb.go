// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.19.6
// source: service_auth copy.proto

package pb

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

const (
	Menu_GetMenu_FullMethodName = "/pb.Menu/GetMenu"
)

// MenuClient is the client API for Menu service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MenuClient interface {
	// register new user
	GetMenu(ctx context.Context, in *GetMenuRequest, opts ...grpc.CallOption) (*GetMenuResponse, error)
}

type menuClient struct {
	cc grpc.ClientConnInterface
}

func NewMenuClient(cc grpc.ClientConnInterface) MenuClient {
	return &menuClient{cc}
}

func (c *menuClient) GetMenu(ctx context.Context, in *GetMenuRequest, opts ...grpc.CallOption) (*GetMenuResponse, error) {
	out := new(GetMenuResponse)
	err := c.cc.Invoke(ctx, Menu_GetMenu_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MenuServer is the server API for Menu service.
// All implementations must embed UnimplementedMenuServer
// for forward compatibility
type MenuServer interface {
	// register new user
	GetMenu(context.Context, *GetMenuRequest) (*GetMenuResponse, error)
	mustEmbedUnimplementedMenuServer()
}

// UnimplementedMenuServer must be embedded to have forward compatible implementations.
type UnimplementedMenuServer struct {
}

func (UnimplementedMenuServer) GetMenu(context.Context, *GetMenuRequest) (*GetMenuResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMenu not implemented")
}
func (UnimplementedMenuServer) mustEmbedUnimplementedMenuServer() {}

// UnsafeMenuServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MenuServer will
// result in compilation errors.
type UnsafeMenuServer interface {
	mustEmbedUnimplementedMenuServer()
}

func RegisterMenuServer(s grpc.ServiceRegistrar, srv MenuServer) {
	s.RegisterService(&Menu_ServiceDesc, srv)
}

func _Menu_GetMenu_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMenuRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MenuServer).GetMenu(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Menu_GetMenu_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MenuServer).GetMenu(ctx, req.(*GetMenuRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Menu_ServiceDesc is the grpc.ServiceDesc for Menu service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Menu_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Menu",
	HandlerType: (*MenuServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetMenu",
			Handler:    _Menu_GetMenu_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service_auth copy.proto",
}
