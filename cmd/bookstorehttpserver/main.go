package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/gorilla/mux"
	pb "github.com/lookuptable/istio-traffic-management-study/pkg/apis/bookstore"
	"github.com/lookuptable/istio-traffic-management-study/pkg/bookstore"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	port = flag.Int("port", 8080, "port")

	server = bookstore.NewServer()
)

func main() {
	flag.Parse()

	r := mux.NewRouter()

	r.HandleFunc("/v1/shelves", createShelf).Methods("POST")
	r.HandleFunc("/v1/shelves", listShelves).Methods("GET")
	r.HandleFunc("/v1/shelves/{id}", deleteShelf).Methods("DELETE")

	http.ListenAndServe(fmt.Sprintf(":%d", *port), r)
}

func createShelf(w http.ResponseWriter, req *http.Request) {
	log.Printf("create shelf")

	if req.Body == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("empty body"))
		return
	}
	defer req.Body.Close()

	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("read body: %v", err)))
		return
	}

	s := &pb.Shelf{}
	if err := protojson.Unmarshal(bytes, s); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Unmarshal: %v", err)))
		return
	}

	cbr := &pb.CreateShelfRequest{Shelf: s}
	cs, err := server.CreateShelf(context.TODO(), cbr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("create shelf: %v", err)))
		return
	}

	bs, err := protojson.Marshal(cs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("marshal: %v", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(bs)
}

func listShelves(w http.ResponseWriter, req *http.Request) {
	log.Printf("list shelves")

	resp, err := server.ListShelves(context.Background(), &empty.Empty{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("list shelves: %v", err)))
		return
	}

	bs, err := protojson.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("marshal: %v", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(bs)
}

func deleteShelf(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	sid, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("shelf ID is required"))
		return
	}

	id, err := strconv.ParseInt(sid, 10, 64)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid shelf ID %q: %v", sid, err)))
		return
	}

	log.Printf("delete shelf")

	dsr := &pb.DeleteShelfRequest{Shelf: id}
	if _, err := server.DeleteShelf(context.TODO(), dsr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("delete shelf: %v", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
}
