package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/micro/protobuf/ptypes"
	v1 "github.com/ryan0906/Memos/pkg/api/v1"
	"google.golang.org/grpc"
)

const (
	apiVersion = "v1"
)

func main() {
	addr := flag.String("server", "", "gRPC server in format host:port")
	flag.Parse()

	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot connect to: %v", err)
	}
	defer conn.Close()

	c := v1.NewMemoServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t := time.Now().In(time.UTC)
	reminder, _ := ptypes.TimestampProto(t)
	pfx := t.Format(time.RFC3339Nano)

	// Create request
	req1 := v1.CreateMemoRequest{
		Api: apiVersion,
		Memo: &v1.Memo{
			Name:        "Memo sample",
			Description: pfx,
			Reminder:    reminder,
		},
	}

	res1, err := c.Create(ctx, &req1)
	if err != nil {
		log.Fatalf("Create failed, error: %v", err)
	}
	log.Printf("Create result: <%v>\n", res1)
	id := res1.Id

	// Read request
	req2 := v1.ReadMemoRequest{
		Api: apiVersion,
		Id:  id,
	}
	res2, err := c.Read(ctx, &req2)
	if err != nil {
		log.Fatalf("Read failed, error: %v", err)
	}
	log.Printf("Read result: <%v>\n", res2)

	// Update request
	req3 := v1.UpdateMemoRequest{
		Api: apiVersion,
		Memo: &v1.Memo{
			Id:          res2.Memo.Id,
			Name:        res2.Memo.Name,
			Description: res2.Memo.Description + "updated",
			Reminder:    res2.Memo.Reminder,
		},
	}
	res3, err := c.Update(ctx, &req3)
	if err != nil {
		log.Fatalf("Update failed, error: %v", err)
	}
	log.Printf("Update result: <%v>\n", res3)

	// Read All request
	req4 := v1.ReadAllRequest{
		Api: apiVersion,
	}
	res4, err := c.ReadAll(ctx, &req4)
	if err != nil {
		log.Fatalf("Read All failed, error: %v", err)
	}
	log.Printf("Read All result: <%v>", res4)

	// Delete request
	req5 := v1.DeleteMemoRequest{
		Api: apiVersion,
		Id:  id,
	}
	res5, err := c.Delete(ctx, &req5)
	if err != nil {
		log.Fatalf("Delete failed, error: %v", err)
	}
	log.Printf("Delete result: <%v>", res5)
}
