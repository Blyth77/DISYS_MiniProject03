package server

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"
)

type Server struct {
	protos.UnimplementedAuctionhouseServiceServer
	ID                   int32
	currentHighestBidder HighestBidder
}

type HighestBidder struct {
	HighestBidAmount int32
	HighestBidderID  int32
	streamBid        protos.AuctionhouseService_BidServer
}

func Start(id int32, po int32) {
	port := po

	s := &Server{}
	s.currentHighestBidder = HighestBidder{}

	s.ID = id

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

	logger.InfoLogger.Printf("Replica%v ready for requests on port: %v\n", s.ID, port)
	output(fmt.Sprintf("Replica%v connected on port: %v", s.ID, port))

	bl := make(chan bool)
	<-bl
}

// BID
func (s *Server) Bid(stream protos.AuctionhouseService_BidServer) error {
	fin := make(chan bool)

	go s.RecieveBidFromClient(fin, stream)

	bl := make(chan error)
	return <-bl
}

func (s *Server) RecieveBidFromClient(fin chan (bool), srv protos.AuctionhouseService_BidServer) {
	for {
		var bid, err = srv.Recv()
		if err != nil {
			logger.ErrorLogger.Printf("FATAL: failed to recive bid from frontend: %s", err)
		} else {
			//Handle new bid - is bid higher than the last highest bid?
			if bid.Amount > s.currentHighestBidder.HighestBidAmount {
				highBidder := HighestBidder{
					HighestBidAmount: bid.Amount,
					HighestBidderID:  bid.ClientId,
					streamBid:        srv,
				}
				s.currentHighestBidder = highBidder
				logger.InfoLogger.Printf("Recived new bid: %d from client%d", bid.Amount, bid.ClientId)
			}
			s.SendBidStatusToClient(srv, bid.ClientId, bid.Amount)
		}
		sleep()
	}
}

func (s *Server) SendBidStatusToClient(srv protos.AuctionhouseService_BidServer, currentBidderID int32, currentBid int32) {
	var status protos.Status
	logger.InfoLogger.Printf("Sending response to frontend%d with %v", currentBidderID, status)

	switch {
	case s.currentHighestBidder.HighestBidderID == currentBidderID && s.currentHighestBidder.HighestBidAmount == currentBid:
		status = protos.Status_NOW_HIGHEST_BIDDER
	case s.currentHighestBidder.HighestBidderID != currentBidderID || s.currentHighestBidder.HighestBidAmount > currentBid:
		status = protos.Status_TOO_LOW_BID
	default:
		status = protos.Status_EXCEPTION
	}

	bidStatus := &protos.StatusOfBid{
		Status: status,
	}

	srv.Send(bidStatus)
	logger.InfoLogger.Printf("Responded to bid from frontend%d.", currentBidderID)
}

// RESULT
func (s *Server) Result(stream protos.AuctionhouseService_ResultServer) error {
	er := make(chan error)

	//go s.receiveQueryFromFrontEndAndSendResponse(stream, er)

	return <-er
}

func (s *Server) receiveQueryFromFrontEndAndSendResponse(srv protos.AuctionhouseService_ResultServer, er_ chan error) {
	for {
		_, err := srv.Recv()
		if err != nil {
			logger.WarningLogger.Printf("FATAL: failed to recive QueryResult from Replica: %s", err)
		} else {

			queryResponse := &protos.ResponseToQuery{
				AuctionStatusMessage: "",
				HighestBid:           s.currentHighestBidder.HighestBidAmount,
				HighestBidderID:      s.currentHighestBidder.HighestBidderID,
				Item:                 "",
			}
			er := srv.Send(queryResponse)
			if er != nil {
				logger.ErrorLogger.Fatalf("FATAL: failed to send ResponseToQuery to frontend: %s", err)
			}
			logger.InfoLogger.Println("Query sent to frontend")
		}
		sleep()
	}
}

// Extensions
func output(input string) {
	fmt.Println(input)
}

func sleep() {
	time.Sleep(2 * time.Second)
}
