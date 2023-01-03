package bookstore

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/lookuptable/istio-traffic-management-study/pkg/apis/bookstore"
)

func NewServer() pb.BookstoreServer {
	return &server{
		Shelves: map[int64]*pb.Shelf{},
		Books:   map[int64]map[int64]*pb.Book{},
	}
}

// The Service type implements a bookstore server.
// All objects are managed in an in-memory non-persistent store.
//
// server is used to implement Bookstoreserver.
type server struct {
	// shelves are stored in a map keyed by shelf id
	// books are stored in a two level map, keyed first by shelf id and then by book id
	Shelves     map[int64]*pb.Shelf
	Books       map[int64]map[int64]*pb.Book
	LastShelfID int64      // the id of the last shelf that was added
	LastBookID  int64      // the id of the last book that was added
	Mutex       sync.Mutex // global mutex to synchronize service access
}

func (s *server) ListShelves(context.Context, *empty.Empty) (*pb.ListShelvesResponse, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// copy shelf ids from Shelves map keys
	shelves := make([]*pb.Shelf, 0, len(s.Shelves))
	for _, shelf := range s.Shelves {
		shelves = append(shelves, shelf)
	}
	resp := &pb.ListShelvesResponse{
		Shelves: shelves,
	}
	return resp, nil
}

func (s *server) CreateShelf(ctx context.Context, req *pb.CreateShelfRequest) (*pb.Shelf, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// assign an id and name to a shelf and add it to the Shelves map.
	shelf := req.Shelf
	s.LastShelfID++
	sid := s.LastShelfID
	s.Shelves[sid] = shelf

	shelf.Id = sid
	return shelf, nil

}

func (s *server) GetShelf(ctx context.Context, req *pb.GetShelfRequest) (*pb.Shelf, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// look up a shelf from the Shelves map.
	shelf, err := s.getShelf(req.Shelf)
	if err != nil {
		return nil, err
	}

	return shelf, nil
}

func (s *server) DeleteShelf(ctx context.Context, req *pb.DeleteShelfRequest) (*empty.Empty, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// delete a shelf by removing the shelf from the Shelves map and the associated books from the Books map.
	delete(s.Shelves, req.Shelf)
	delete(s.Books, req.Shelf)
	return &empty.Empty{}, nil
}

func (s *server) ListBooks(ctx context.Context, req *pb.ListBooksRequest) (*pb.ListBooksResponse, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// list the books in a shelf
	if _, err := s.getShelf(req.Shelf); err != nil {
		return nil, err
	}
	shelfBooks := s.Books[req.Shelf]
	books := make([]*pb.Book, 0, len(shelfBooks))
	for _, book := range shelfBooks {
		books = append(books, book)
	}

	resp := &pb.ListBooksResponse{
		Books: books,
	}
	return resp, nil
}

func (s *server) CreateBook(ctx context.Context, req *pb.CreateBookRequest) (*pb.Book, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	_, err := s.getShelf(req.Shelf)
	if err != nil {
		return nil, err
	}
	// assign an id and name to a book and add it to the Books map.
	s.LastBookID++
	bid := s.LastBookID
	book := req.Book
	if s.Books[req.Shelf] == nil {
		s.Books[req.Shelf] = make(map[int64]*pb.Book)
	}
	s.Books[req.Shelf][bid] = book

	return book, nil
}

func (s *server) GetBook(ctx context.Context, req *pb.GetBookRequest) (*pb.Book, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// get a book from the Books map
	book, err := s.getBook(req.Shelf, req.Book)
	if err != nil {
		return nil, err
	}

	return book, nil
}

func (s *server) DeleteBook(ctx context.Context, req *pb.DeleteBookRequest) (*empty.Empty, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// delete a book by removing the book from the Books map.
	delete(s.Books[req.Shelf], req.Book)
	return nil, nil
}

// internal helpers
func (s *server) getShelf(sid int64) (shelf *pb.Shelf, err error) {
	shelf, ok := s.Shelves[sid]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Couldn't find shelf %d", sid))
	} else {
		return shelf, nil
	}
}

func (s *server) getBook(sid int64, bid int64) (book *pb.Book, err error) {
	_, err = s.getShelf(sid)
	if err != nil {
		return nil, err
	}
	book, ok := s.Books[sid][bid]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Couldn't find book %d on shelf %d", bid, sid))
	} else {
		return book, nil
	}
}
