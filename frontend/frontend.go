package server

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"

	"google.golang.org/grpc"
)

var (
	ID                   int32
	currentHighestBidder = HighestBidder{}
	connected            bool
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

type frontendClientReplica struct {
	clientService protos.AuctionhouseServiceClient
	conn          *grpc.ClientConn
}

type FrontendClienthandle struct {
	streamBidOut    protos.AuctionhouseService_BidClient
	streamResultOut protos.AuctionhouseService_ResultClient
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
		frontendClientForReplica := setupFrontend(po)

		channelBid := frontendClientForReplica.setupBidStream()
		channelResult := frontendClientForReplica.setupResultStream()

		go channelResult.receiveFromResultStream()
		go channelBid.recvBidStatus()

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
// skal ikke store auctioneers
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
// recieve from replica
func (s *Server) SendBidStatusToClient(stream protos.AuctionhouseService_BidServer, currentBidderID int32, currentBid int32) {
	var status protos.Status

	// Her skal den svarer clienten NÅR!!!!! den har modtaget majority af ackno fra replicas!
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

// Client To replica del

func (client *frontendClientReplica) setupBidStream() FrontendClienthandle {
	streamOut, err := client.clientService.Bid(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to call AuctionhouseService: %v", err)
	}
	return FrontendClienthandle{streamBidOut: streamOut}
}

func (frontendClient *frontendClientReplica) setupResultStream() FrontendClienthandle {
	streamOut, err := frontendClient.clientService.Result(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to call AuctionhouseService: %v", err)
	}

	return FrontendClienthandle{streamResultOut: streamOut}
}

func setupFrontend(port string) *frontendClientReplica {
	setupFClientReplicaID()

	logger.LogFileInit("client", ID)

	frontendClient, err := makeFrontendClient(port)
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to make Client: %v", err)
	}

	return frontendClient
}

func setupFClientReplicaID() {
	rand.Seed(time.Now().UnixNano())
	ID = int32(rand.Intn(1e4))
}

// When client has sent a bid request - recieves a status message: success, fail or expection
func (frontendCh *FrontendClienthandle) recvBidStatus() {
	for {
		msg, err := frontendCh.streamBidOut.Recv()
		if err != nil {
			logger.ErrorLogger.Printf("Error in receiving message from server: %v", msg)
			connected = false
			time.Sleep(5 * time.Second) // waiting before trying to recieve again
		} else {
			// FRONTEND: skal vente på majority har acknowledged og svaret, før den godtager at de har gemt bid.
			switch msg.Status {
			case protos.Status_NOW_HIGHEST_BIDDER:
			case protos.Status_TOO_LOW_BID:
			case protos.Status_EXCEPTION:
			}
			connected = true
		}
	}
}

// should this then wait for majority and send the result to the client?
func (frontendCh *FrontendClienthandle) receiveFromResultStream() {
	for {
		if !connected { // To avoid sending before connected.
			time.Sleep(1 * time.Second)
		} else {
			response, err := frontendCh.streamResultOut.Recv()
			if err != nil {
				logger.ErrorLogger.Printf("Failed to receive message: %v", err)
			} else {
				Output(fmt.Sprintf("Current highest bid: %v from clientID: %v", response.HighestBid, response.HighestBidderID))
				logger.InfoLogger.Println("Succesfully recieved response from query")
			}
		}
	}
}

//Connects and creates client through protos.NewAuctionhouseServiceClient(connection)
func makeFrontendClient(port string) (*frontendClientReplica, error) {

	conn, err := makeConnection(port)
	if err != nil {
		return nil, err
	}

	return &frontendClientReplica{
		clientService: protos.NewAuctionhouseServiceClient(conn),
		conn:          conn,
	}, nil
}

func makeConnection(port string) (*grpc.ClientConn, error) {
	logger.InfoLogger.Print("Connecting to the auctionhouse...")
	return grpc.Dial(port, []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}...)
}

func Output(input string) {
	fmt.Println(input)
}
