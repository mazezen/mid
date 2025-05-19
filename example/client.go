package main

import (
	"context"
	"log"
	"time"

	"github.com/mazezen/mid/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func main() {
	conn, err := grpc.Dial(
		"localhost:50051", 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
   		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // 发送 ping 的间隔
			Timeout:             10 * time.Second,  // 等待 ping 响应的超时
			PermitWithoutStream: true,             // 允许无活跃流时发送 ping
		}),
	)
	if err != nil {
		log.Fatalf("failed to connect :%v", err)
	}
	defer conn.Close()

	client := pb.NewIDMakerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// 测试 Snowflake 模式
	resp, err := client.MakeIDService(ctx, &pb.MakeIDServiceRequest{Mode: "snowflake"})
	if err != nil {
		log.Fatalf("failed to generate snowflake ID: %v", err)
	}
	log.Printf("Snowflake ID: %d", resp.Id)

	// 测试 Segment 模式
	resp, err = client.MakeIDService(ctx, &pb.MakeIDServiceRequest{Mode: "segment"})
	if err != nil {
		log.Fatalf("failed to generate segment ID: %v", err)
	}
	log.Printf("Segment ID: %d", resp.Id)
}