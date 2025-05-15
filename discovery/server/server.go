package main

import (
	"context"
	"etcd/discovery"
	pb "etcd/discovery/proto"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
)

var (
	// 定义命令行参数 port，默认值为 50051)
	port = flag.Int("port", 50051, "")
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("Recv Client : %v \n", in.Msg)
	return &pb.HelloReply{Msg: "Hello Client"}, nil
}

func (s *server) SayHi(ctx context.Context, in *pb.HiRequest) (*pb.HiReply, error) {
	fmt.Printf("Recv Client : %v \n", in.Msg)
	return &pb.HiReply{Msg: "Hi Client"}, nil
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 创建 TCP 监听，绑定到指定端口
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v\n", err)
	}

	// 创建 gRPC 服务器实例
	s := grpc.NewServer()

	// 注册服务实现到 gRPC 服务器
	serverRegister(s, &server{})

	// 启动 gRPC 服务器并监听连接
	if err = s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v\n", err)
	}
}

func serverRegister(s grpc.ServiceRegistrar, srv pb.GreeterServer) {
	pb.RegisterGreeterServer(s, srv)
	s1 := &discovery.Service{
		Name:     "hello.Greeter",
		IP:       "localhost",
		Port:     strconv.Itoa(*port),
		Protocol: "grpc",
	}
	go discovery.ServiceRegister(s1)
}
