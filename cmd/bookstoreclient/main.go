package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"
	pb "github.com/lookuptable/istio-traffic-management-study/pkg/apis/bookstore"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var server = flag.String("server", "localhost:8080", "server address")

const (
	shelfTheme = "friction"
)

func main() {
	flag.Parse()

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(*server, opt)
	if err != nil {
		log.Fatalf("create client connection: %v", err)
	}
	defer conn.Close()

	c := pb.NewBookstoreClient(conn)
	ctx := context.Background()

	shelf, err := createShelf(ctx, c)
	if err != nil {
		log.Fatal(err)
	}

	if err := createBooks(ctx, c, shelf.GetId()); err != nil {
		log.Fatal(err)
	}

	if err := listBooks(ctx, c, shelf.GetId()); err != nil {
		log.Fatal(err)
	}

	if err := deleteShelf(ctx, c, shelf.GetId()); err != nil {
		log.Fatal(err)
	}
}

func listBooks(ctx context.Context, c pb.BookstoreClient, shelfID int64) error {
	lbr := &pb.ListBooksRequest{Shelf: shelfID}
	resp, err := c.ListBooks(ctx, lbr)
	if err != nil {
		return fmt.Errorf("list books: %v", err)
	}
	for _, book := range resp.Books {
		log.Printf("%s", proto.MarshalTextString(book))
	}
	return nil
}

func createBooks(ctx context.Context, c pb.BookstoreClient, shelfID int64) error {
	for i := 0; i < 10; i++ {
		cbr := &pb.CreateBookRequest{
			Shelf: shelfID,
			Book: &pb.Book{
				Id:     int64(i),
				Author: fmt.Sprintf("author-%d", i),
				Title:  fmt.Sprintf("title-%d", i),
			},
		}
		if _, err := c.CreateBook(ctx, cbr); err != nil {
			return fmt.Errorf("create book: %v", err)
		}
	}
	log.Printf("created books")
	return nil
}

func createShelf(ctx context.Context, c pb.BookstoreClient) (*pb.Shelf, error) {
	csr := &pb.CreateShelfRequest{
		Shelf: &pb.Shelf{
			Theme: shelfTheme,
		},
	}
	shelf, err := c.CreateShelf(ctx, csr)
	if err != nil {
		return nil, fmt.Errorf("create shelf: %v", err)
	}

	log.Printf("created shelf: %v", shelf)
	return shelf, nil
}

func deleteShelf(ctx context.Context, c pb.BookstoreClient, shelfID int64) error {
	dsr := &pb.DeleteShelfRequest{Shelf: shelfID}
	if _, err := c.DeleteShelf(ctx, dsr); err != nil {
		return fmt.Errorf("delete shelf: %v", err)
	}
	log.Printf("deleted shelf - ID %d", shelfID)
	return nil
}
