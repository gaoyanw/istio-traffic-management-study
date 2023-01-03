package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	pb "github.com/lookuptable/istio-traffic-management-study/pkg/apis/bookstore"
	"github.com/lookuptable/istio-traffic-management-study/pkg/bookstore"
	"google.golang.org/grpc"
	klog "k8s.io/klog/v2"
)

var port = flag.Int("port", 8080, "port number")

func main() {
	klog.InitFlags(nil)
	klog.SetOutput(os.Stderr)

	flag.Parse()

	RunServer()
}

func RunServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		klog.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	fmt.Printf("\nServer listening on port %d \n", *port)
	pb.RegisterBookstoreServer(s, bookstore.NewServer())
	if err := s.Serve(lis); err != nil {
		klog.Fatalf("failed to serve: %v", err)
	}
}
