package main

import (
	"fmt"
	"net"
	"os"
	"sync"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"

	"google.golang.org/grpc"
)

var (
	serverId             int32
	port                 = 3000
	currentHighestBidder = HighestBidder{}
)

type Server struct {
	protos.UnimplementedAuctionhouseServiceServer
	auctioneer  sync.Map
	unsubscribe []int32
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

func main() {
	serverId = 1 // Unhardcode : must get it from main
	logger.LogFileInit("server", serverId)

	s := &Server{}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.InfoLogger.Printf(fmt.Sprintf("FATAL: Connection unable to establish. Failed to listen: %v", err))
	}

	grpcServer := grpc.NewServer()

	protos.RegisterAuctionhouseServiceServer(grpcServer, s)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.ErrorLogger.Fatalf("FATAL: Server connection failed: %s", err)
		}
	}()
	Output(fmt.Sprintf("Server connected on port: %v", port))

	var o string
	fmt.Scanln(&o)
	os.Exit(3)
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
			logger.ErrorLogger.Println(fmt.Sprintf("FATAL: failed to recive bid from client: %s", err))
		} else {
			//check if client is subscribed
			_, ok := s.auctioneer.Load(bid.ClientId)
			if !ok {
				s.auctioneer.Store(bid.ClientId, sub{streamBid: srv, finished: fin})
				logger.InfoLogger.Printf("Storing new client %v, in server map", bid.ClientId)
			}

			//Handle new bid - is bid higher than the last highest bid?
			if bid.Amount > currentHighestBidder.HighestBidAmount {
				highBidder := HighestBidder{
					HighestBidAmount: bid.Amount,
					HighestBidderID:  bid.ClientId,
					streamBid:        srv,
				}
				currentHighestBidder = highBidder
				logger.InfoLogger.Printf("Storing new bid %d for client %d in server map", bid.Amount, bid.ClientId)
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

// Unsubscribes client
func (s *Server) killAuctioneer() {
	for _, id := range s.unsubscribe {
		//logger.InfoLogger.Printf("Killed client: %v", id)

		idd := int32(id)
		m, ok := s.auctioneer.Load(idd)
		if !ok && m != nil {
			logger.InfoLogger.Println(fmt.Sprintf("Failed to find auctioneer id: %T", idd))
		}
		sub, ok := m.(sub)
		if !ok && m != nil {
			logger.WarningLogger.Panicf("Failed to cast auctioneer id: %T", sub)
		}
		if m != nil {
			Output(fmt.Sprintf("Client%v left the auction", idd))
		}
		s.auctioneer.Delete(id)
		logger.InfoLogger.Printf("Client id %v has been killed and deleted", id)
	}
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
			logger.WarningLogger.Printf("FATAL: failed to recive QueryResult from client: %s", err)
		} else {

			queryResponse := &protos.ResponseToQuery{
				AuctionStatusMessage: "",
				HighestBid:           currentHighestBidder.HighestBidAmount,
				HighestBidderID:      currentHighestBidder.HighestBidderID,
				Item:                 "",
			}
			er := srv.Send(queryResponse)
			if er != nil {
				logger.ErrorLogger.Fatalf("FATAL: failed to send ResponseToQuery to client: %s", err)
			}
			logger.InfoLogger.Println("Query sent to client")
		}
	}
}



func Output(input string) {
	fmt.Println(input)
}
