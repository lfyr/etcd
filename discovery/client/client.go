package main

import (
	"context"
	"etcd/discovery"
	pb "etcd/discovery/proto"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func getServerAddr(svcName string) string {
	s := discovery.ServiceDiscovery(svcName)
	if s == nil {
		return ""
	}
	if s.IP == "" && s.Port == "" {
		return ""
	}
	return s.IP + ":" + s.Port
}

func sayHello() {
	addr := getServerAddr("hello.Greeter")
	if addr == "" {
		log.Println("未发现可用服务")
		return
	}
	fmt.Println(addr)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("did not connect: %v", err)
		return
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)
	r, err := c.SayHello(context.Background(), &pb.HelloRequest{Msg: "Hello Server"})
	if err != nil {
		log.Println("could not greet: %v", err)
		return
	}
	log.Printf("Recv Server : %s", r.Msg)

	h, err := c.SayHi(context.Background(), &pb.HiRequest{Msg: "Hi Server"})
	if err != nil {
		log.Println("could not greet: %v", err)
		return
	}
	log.Printf("Recv Server : %s", h.Msg)
}

func main() {
	log.SetFlags(log.Llongfile)
	go discovery.WatchServiceName("hello.Greeter")
	for {
		sayHello()
		time.Sleep(time.Second * 2)
	}
}
