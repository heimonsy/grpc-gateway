package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var (
	director proxy.StreamDirector
	conn     *grpc.ClientConn
)

func main() {
	server := grpc.NewServer(grpc.CustomCodec(proxy.Codec()))
	do_proxy()
	proxy.RegisterService(
		server,
		director,
		"heimonsy.grpc.Example",
		"Add",
		"Connect",
	)

	ln, err := net.Listen("tcp", ":8810")
	if err != nil {
		log.Fatalln(err)
	}
	err = server.Serve(ln)
	if err != nil {
		log.Fatalln("serve error", err)
	}

}

func do_proxy() {
	conn, err := grpc.DialContext(
		context.Background(),
		"127.0.0.1:8811",
		grpc.WithCodec(proxy.Codec()),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatal(err)
	}

	director = func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		// Make sure we never forward internal services.
		//if strings.HasPrefix(fullMethodName, "/com.example.internal.") {
		//return nil, nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
		//}

		md, ok := metadata.FromIncomingContext(ctx)
		fmt.Println(md)
		// Copy the inbound metadata explicitly.
		outCtx, _ := context.WithCancel(ctx)
		outCtx = metadata.NewOutgoingContext(outCtx, md.Copy())
		if ok {
			// Decide on which backend to dial
			//if val, exists := md[":authority"]; exists && val[0] == "staging.api.example.com" {
			//// Make sure we use DialContext so the dialing can be cancelled/time out together with the context.
			//conn, err := grpc.DialContext(ctx, "api-service.staging.svc.local", grpc.WithCodec(proxy.Codec()))
			//return outCtx, conn, err
			//} else if val, exists := md[":authority"]; exists && val[0] == "api.example.com" {
			//conn, err := grpc.DialContext(ctx, "api-service.prod.svc.local", grpc.WithCodec(proxy.Codec()))
			//return outCtx, conn, err
			//}
			//conn, err := grpc.DialContext(
			//ctx,
			//"127.0.0.1:8811",
			//grpc.WithCodec(proxy.Codec()),
			//grpc.WithInsecure(),
			//)
			return outCtx, conn, err
		}
		return nil, nil, grpc.Errorf(codes.Internal, "Can't find metadata")
	}
}
