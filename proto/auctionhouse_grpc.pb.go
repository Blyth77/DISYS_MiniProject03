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

// AuctionhouseServiceClient is the client API for AuctionhouseService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuctionhouseServiceClient interface {
	Result(ctx context.Context, opts ...grpc.CallOption) (AuctionhouseService_ResultClient, error)
	Bid(ctx context.Context, opts ...grpc.CallOption) (AuctionhouseService_BidClient, error)
}

type auctionhouseServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuctionhouseServiceClient(cc grpc.ClientConnInterface) AuctionhouseServiceClient {
	return &auctionhouseServiceClient{cc}
}

func (c *auctionhouseServiceClient) Result(ctx context.Context, opts ...grpc.CallOption) (AuctionhouseService_ResultClient, error) {
	stream, err := c.cc.NewStream(ctx, &AuctionhouseService_ServiceDesc.Streams[0], "/proto.AuctionhouseService/Result", opts...)
	if err != nil {
		return nil, err
	}
	x := &auctionhouseServiceResultClient{stream}
	return x, nil
}

type AuctionhouseService_ResultClient interface {
	Send(*QueryResult) error
	Recv() (*ResponseToQuery, error)
	grpc.ClientStream
}

type auctionhouseServiceResultClient struct {
	grpc.ClientStream
}

func (x *auctionhouseServiceResultClient) Send(m *QueryResult) error {
	return x.ClientStream.SendMsg(m)
}

func (x *auctionhouseServiceResultClient) Recv() (*ResponseToQuery, error) {
	m := new(ResponseToQuery)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *auctionhouseServiceClient) Bid(ctx context.Context, opts ...grpc.CallOption) (AuctionhouseService_BidClient, error) {
	stream, err := c.cc.NewStream(ctx, &AuctionhouseService_ServiceDesc.Streams[1], "/proto.AuctionhouseService/Bid", opts...)
	if err != nil {
		return nil, err
	}
	x := &auctionhouseServiceBidClient{stream}
	return x, nil
}

type AuctionhouseService_BidClient interface {
	Send(*BidRequest) error
	Recv() (*StatusOfBid, error)
	grpc.ClientStream
}

type auctionhouseServiceBidClient struct {
	grpc.ClientStream
}

func (x *auctionhouseServiceBidClient) Send(m *BidRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *auctionhouseServiceBidClient) Recv() (*StatusOfBid, error) {
	m := new(StatusOfBid)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// AuctionhouseServiceServer is the server API for AuctionhouseService service.
// All implementations must embed UnimplementedAuctionhouseServiceServer
// for forward compatibility
type AuctionhouseServiceServer interface {
	Result(AuctionhouseService_ResultServer) error
	Bid(AuctionhouseService_BidServer) error
	mustEmbedUnimplementedAuctionhouseServiceServer()
}

// UnimplementedAuctionhouseServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAuctionhouseServiceServer struct {
}

func (UnimplementedAuctionhouseServiceServer) Result(AuctionhouseService_ResultServer) error {
	return status.Errorf(codes.Unimplemented, "method Result not implemented")
}
func (UnimplementedAuctionhouseServiceServer) Bid(AuctionhouseService_BidServer) error {
	return status.Errorf(codes.Unimplemented, "method Bid not implemented")
}
func (UnimplementedAuctionhouseServiceServer) mustEmbedUnimplementedAuctionhouseServiceServer() {}

// UnsafeAuctionhouseServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuctionhouseServiceServer will
// result in compilation errors.
type UnsafeAuctionhouseServiceServer interface {
	mustEmbedUnimplementedAuctionhouseServiceServer()
}

func RegisterAuctionhouseServiceServer(s grpc.ServiceRegistrar, srv AuctionhouseServiceServer) {
	s.RegisterService(&AuctionhouseService_ServiceDesc, srv)
}

func _AuctionhouseService_Result_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(AuctionhouseServiceServer).Result(&auctionhouseServiceResultServer{stream})
}

type AuctionhouseService_ResultServer interface {
	Send(*ResponseToQuery) error
	Recv() (*QueryResult, error)
	grpc.ServerStream
}

type auctionhouseServiceResultServer struct {
	grpc.ServerStream
}

func (x *auctionhouseServiceResultServer) Send(m *ResponseToQuery) error {
	return x.ServerStream.SendMsg(m)
}

func (x *auctionhouseServiceResultServer) Recv() (*QueryResult, error) {
	m := new(QueryResult)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _AuctionhouseService_Bid_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(AuctionhouseServiceServer).Bid(&auctionhouseServiceBidServer{stream})
}

type AuctionhouseService_BidServer interface {
	Send(*StatusOfBid) error
	Recv() (*BidRequest, error)
	grpc.ServerStream
}

type auctionhouseServiceBidServer struct {
	grpc.ServerStream
}

func (x *auctionhouseServiceBidServer) Send(m *StatusOfBid) error {
	return x.ServerStream.SendMsg(m)
}

func (x *auctionhouseServiceBidServer) Recv() (*BidRequest, error) {
	m := new(BidRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// AuctionhouseService_ServiceDesc is the grpc.ServiceDesc for AuctionhouseService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuctionhouseService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.AuctionhouseService",
	HandlerType: (*AuctionhouseServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Result",
			Handler:       _AuctionhouseService_Result_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "Bid",
			Handler:       _AuctionhouseService_Bid_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "auctionhouse.proto",
}
