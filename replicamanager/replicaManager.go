package server

import (
	"fmt"
	"net"
	"sync"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"

	"google.golang.org/grpc"
)

var (
	ID                   int32
	currentHighestBidder = HighestBidder{}
)

type Server struct {
	protos.UnimplementedAuctionhouseServiceServer
	auctioneer  sync.Map
}

type sub struct {
	streamBid protos.AuctionhouseService_BidServer
	finished  chan<- bool
}

type HighestBidder struct {
	HighestBidAmount int32
	HighestBidderID  int32
	streamBid        protos.AuctionhouseService_BidServer
}

func Start(id int32, po int32) {
	port := po
	logger.LogFileInit("replica", id)

	s := &Server{}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.InfoLogger.Printf(fmt.Sprintf("FATAL: Connection unable to establish. Failed to listen: %v", err))
	}

	grpcServer := grpc.NewServer()

	protos.RegisterAuctionhouseServiceServer(grpcServer, s)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.ErrorLogger.Fatalf("FATAL: replica connection failed: %s", err)
		}
	}()
	logger.InfoLogger.Println("mis")
	Output(fmt.Sprintf("Replica connected on port: %v", port))

	bl := make(chan bool)
	<-bl
}

func (s *Server) Bid(stream protos.AuctionhouseService_BidServer) error {
	fin := make(chan bool)

	go s.HandleNewBidForClient(fin, stream)

	bl := make(chan error)
	return <-bl
}

func (s *Server) HandleNewBidForClient(fin chan (bool), srv protos.AuctionhouseService_BidServer) {
	for {
		var bid, err = srv.Recv()
		if err != nil {
			logger.ErrorLogger.Println(fmt.Sprintf("FATAL: failed to recive bid from frontend: %s", err))
		} else {
			//check if client is subscribed
			_, ok := s.auctioneer.Load(bid.ClientId)
			if !ok {
				s.auctioneer.Store(bid.ClientId, sub{streamBid: srv, finished: fin})
				logger.InfoLogger.Printf("Storing new client-frontend %v, in Replica map", bid.ClientId)
			}

			//Handle new bid - is bid higher than the last highest bid?
			if bid.Amount > currentHighestBidder.HighestBidAmount {
				highBidder := HighestBidder{
					HighestBidAmount: bid.Amount,
					HighestBidderID:  bid.ClientId,
					streamBid:        srv,
				}
				currentHighestBidder = highBidder
				logger.InfoLogger.Printf("Storing new bid %d for frontend %d in Replica map", bid.Amount, bid.ClientId)
			}
			s.SendBidStatusToClient(srv, bid.ClientId, bid.Amount)
		}
	}
}

func (s *Server) SendBidStatusToClient(stream protos.AuctionhouseService_BidServer, currentBidderID int32, currentBid int32) {
	var status protos.Status

	switch {
	case currentHighestBidder.HighestBidderID == currentBidderID && currentHighestBidder.HighestBidAmount == currentBid:
		status = protos.Status_NOW_HIGHEST_BIDDER
	case currentHighestBidder.HighestBidderID != currentBidderID || currentHighestBidder.HighestBidAmount > currentBid:
		status = protos.Status_TOO_LOW_BID
	default:
		status = protos.Status_EXCEPTION
	}

	bidStatus := &protos.StatusOfBid{
		Status:     status,
		HighestBid: currentHighestBidder.HighestBidAmount,
	}

	stream.Send(bidStatus)
}


// When time has runned out : brodcast who the winner is
func (s *Server) Result(stream protos.AuctionhouseService_ResultServer) error {
	er := make(chan error)

	go s.receiveQueryForResultAndSendToClient(stream, er)

	return <-er
}

// wait for a client to ask for the highest bidder and sends the result back
func (s *Server) receiveQueryForResultAndSendToClient(srv protos.AuctionhouseService_ResultServer, er_ chan error) {
	for {
		_, err := srv.Recv()
		if err != nil {
			logger.WarningLogger.Printf("FATAL: failed to recive QueryResult from Replica: %s", err)
		} else {

			queryResponse := &protos.ResponseToQuery{
				AuctionStatusMessage: "",
				HighestBid:           currentHighestBidder.HighestBidAmount,
				HighestBidderID:      currentHighestBidder.HighestBidderID,
				Item:                 "",
			}
			er := srv.Send(queryResponse)
			if er != nil {
				logger.ErrorLogger.Fatalf("FATAL: failed to send ResponseToQuery to frontend: %s", err)
			}
			logger.InfoLogger.Println("Query sent to frontend")
		}
	}
}

func Output(input string) {
	fmt.Println(input)
}
