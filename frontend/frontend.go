package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
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
	auctioneer sync.Map
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

func Start(id int32, port string) {
	connectToNode(port) //clienten's server og den er correct

	file, _ := os.Open("replicamanager/portlist/listOfReplicaPorts.txt")


	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		scanner.Scan()
		po := scanner.Text()

		// Her skal de 3 metoder ind. og 2 go rutiner 



		// Næste step. Når frontend modtager 5 forskellige svar fra fem replicas, 
		// skal der tages majority og sendes tilbage til clienten. 
		// svarne fra go rutinerne skal samles i en liste og så skal de compares på en eller anden måde. 

	}


	bl := make(chan bool)
	<-bl

}

func connectToNode(port string) {
	s := &Server{}

	lis, err := net.Listen("tcp", port)
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

	bl := make(chan bool)
	<-bl
}

func (s *Server) Bid(stream protos.AuctionhouseService_BidServer) error {
	fin := make(chan bool)

	go s.HandleNewBidForClient(fin, stream)

	bl := make(chan error)
	return <-bl
}

// TODO ... skal modtage fra client of forwarde ti replicas. 
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

//TODO
func (s *Server) SendBidStatusToClient(stream protos.AuctionhouseService_BidServer, currentBidderID int32, currentBid int32) {
	var status protos.Status

	// Her skal den svarer clietn NÅR!!!!! den har modtaget majority af ackno fra replicas!
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
			logger.WarningLogger.Printf("FATAL: failed to recive QueryResult from client: %s", err)
		} else {
			// her skal den finde majority af replicasnes svar
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
