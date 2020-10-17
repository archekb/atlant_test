package main

import (
	"log"
	"net"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"

	"context"

	pb "atlant/productservice/proto"
)

const (
	dbURL   = "mongodb://root:password@db:27017"
	srvPort = ":50051"
)

// productServer is used to implement ProductService. It contains PriceCoder, Fetcher and DB connector
type productServer struct {
	priceCoder *PriceCoder
	fetcher    *Fetcher
	db         *DB
}

// Fetch - implements ProductService.Fetch
func (s *productServer) Fetch(ctx context.Context, in *pb.FetchRequest) (*pb.FetchResponse, error) {
	err := s.fetcher.Fetch(in.GetUrl())
	if err != nil {
		return nil, err
	}

	return &pb.FetchResponse{Success: true}, nil
}

// List - implement ProductService.List
func (s *productServer) List(ctx context.Context, in *pb.ListRequest) (*pb.ListResponse, error) {
	// make sorting parametrs
	sort := []bson.E{}
	for _, s := range in.GetSort() {
		if s.GetDsc() {
			sort = append(sort, bson.E{Key: s.GetSortby().String(), Value: -1})
		} else {
			sort = append(sort, bson.E{Key: s.GetSortby().String(), Value: 1})
		}
	}

	// calculate paging
	resultPerPage := in.GetPaging().GetResultPerPage()
	pageNumber := in.GetPaging().GetPageNumber()
	skip := pageNumber * resultPerPage

	pagesCount, err := s.db.PagesCount(ctx, resultPerPage)
	if err != nil {
		return nil, err
	}

	// get list of sorted products
	res, err := s.db.List(ctx, skip, resultPerPage, sort...)
	if err != nil {
		return nil, err
	}

	var products = make([]*pb.ListResponse_Product, len(res))
	for i, p := range res {
		products[i] = &pb.ListResponse_Product{Name: p.Name, Price: p.Price, LastUpdated: p.LastUpdated.Format(time.RFC3339), ChangesCount: p.ChangesCount}
	}

	return &pb.ListResponse{Paging: &pb.PagingParams{PageNumber: pageNumber, ResultPerPage: resultPerPage, PageCont: pagesCount}, Result: products}, nil
}

func main() {
	// create db connector
	db, err := NewDB(dbURL)
	if err != nil {
		log.Fatalln("Can't connect to MongoDB, by", dbURL, "error:", err)
	}
	defer db.Disconnect()

	// create price coder witn 2 digits after point
	priceCoder := NewPriceCoder(2)

	// create new fetcher with ';' - delimetr, '#' - comment, 2 - fields per line
	fetcher := NewFetcher(';', '#', 2, priceCoder, db)

	// create instance of productService
	pSrv := &productServer{priceCoder: NewPriceCoder(2), fetcher: fetcher, db: db}

	// create listener and start serve
	lis, err := net.Listen("tcp", srvPort)
	if err != nil {
		log.Fatalln("Can't start listener, error:", err)
	}

	grpcSrv := grpc.NewServer()

	pb.RegisterProductServiceServer(grpcSrv, pSrv)
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalln("Can't start serve, error:", err)
	}
}
