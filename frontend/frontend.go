/* TODO:
- FRONTEND: recvBidStatus() - skal vente på majority har acknowledged og svaret, før den godtager at de har gemt bid.
- Hvorfor grr katter den kun 2,5 gange
- HVORFOR KØRER LOOPSNE SÅ MANGE GANGE!!!!
- Når frontend modtager 5 forskellige svar fra fem replicas,
- SendBidStatusToClient: make this
*/

package frontend

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"

	protos "github.com/Blyth77/DISYS_MiniProject03/proto"
	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
)

var (
	ID        int32
	connected bool
)

type Server struct {
	protos.UnimplementedAuctionhouseServiceServer
}

type frontendClientReplica struct {
	clientService protos.AuctionhouseServiceClient
	conn          *grpc.ClientConn
}

type frontendClienthandle struct {
	streamBidOut    protos.AuctionhouseService_BidClient
	streamResultOut protos.AuctionhouseService_ResultClient
}

type message struct {
	Id     int32
	Amount int32
}

type msgQueue struct {
	MessageQue []message
	mu         sync.Mutex
}

var messageHandle = msgQueue{}


func Start(id int32, port string) {

	go connectToClient(port) //clienten's server

	file, _ := os.Open("replicamanager/portlist/listOfReplicaPorts.txt")
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		scanner.Scan()
		po := scanner.Text()

		frontendClientForReplica := setupFrontendConnectionToReplica(po)
		Output(fmt.Sprintf("Frontend connected with replica on port: %v", po))

		channelBid := frontendClientForReplica.setupBidStream()
		channelResult := frontendClientForReplica.setupResultStream()

		go channelBid.recieveBidResponseFromReplicasAndSendToClient()
		go channelResult.recieveQueryResponseFromReplicaAndSendToClient()
		go forwardBidToReplica(channelBid)
	}

	bl := make(chan bool)
	<-bl
}

// CLIENT - BID
func (s *Server) Bid(stream protos.AuctionhouseService_BidServer) error {
	fin := make(chan bool)

	go s.recieveBidRequestFromClient(fin, stream)

	bl := make(chan error)
	return <-bl
}

func (s *Server) recieveBidRequestFromClient(fin chan (bool), srv protos.AuctionhouseService_BidServer) {
	for {
		var bid, err = srv.Recv()
		if err != nil {
			logger.ErrorLogger.Println(fmt.Sprintf("FATAL: failed to recive bid from client: %s", err))
		} else {
			/*
				save srv.recv info in ex an TryClientBid

				send(srv.Context())
			*/
			ID = bid.ClientId

			//check if client is subscribed
			addToMessageQueue(bid.ClientId, bid.Amount)

		}
	}
}

//TODO
func (s *Server) sendBidStatusToClient(stream protos.AuctionhouseService_BidServer, currentBidderID int32, currentBid int32) {
	/*var status protos.Status

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
	*/
}

// CLIENT - RESULT
func (s *Server) Result(stream protos.AuctionhouseService_ResultServer) error {
	er := make(chan error)

	go s.recieveQueryRequestFromClientAndSendToReplica(stream, er)

	return <-er
}

//TODO
func (s *Server) sendResultResponseToClient(stream protos.AuctionhouseService_BidServer, currentBidderID int32, currentBid int32) {
	/*var status protos.Status

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
	*/
}

func (s *Server) recieveQueryRequestFromClientAndSendToReplica(srv protos.AuctionhouseService_ResultServer, er_ chan error) {
	for {
		_, err := srv.Recv()
		if err != nil {
			logger.WarningLogger.Printf("FATAL: failed to recive QueryResult from client: %s", err)
		} else {
			// her skal den finde majority af replicasnes svar
			queryResponse := &protos.ResponseToQuery{
				AuctionStatusMessage: "",
				//HighestBid:           currentHighestBidder.HighestBidAmount,
				//HighestBidderID:      currentHighestBidder.HighestBidderID,
				Item: "",
			}
			er := srv.Send(queryResponse)
			if er != nil {
				logger.ErrorLogger.Fatalf("FATAL: failed to send ResponseToQuery to client: %s", err)
			}
			logger.InfoLogger.Println("Query sent to client")
		}
	}
}

// REPLICA - BID
func forwardBidToReplica(ch frontendClienthandle) {
	for {
		for {
			messageHandle.mu.Lock()

			if len(messageHandle.MessageQue) == 0 {
				messageHandle.mu.Unlock()
				break
			}

			idFromClient := messageHandle.MessageQue[0].Id
			amountFromClient := messageHandle.MessageQue[0].Amount

			messageHandle.mu.Unlock()

			// Send data over the gRPC stream to the client
			if err := ch.streamBidOut.Send(&protos.BidRequest{
				ClientId: idFromClient,
				Amount:   amountFromClient,
			}); err != nil {
				logger.ErrorLogger.Output(2, (fmt.Sprintf("Failed to send data to client: %v", err)))
				// In case of error the client would re-subscribe so close the subscriber stream
			}
			logger.InfoLogger.Println("Forwarding message to replicas..")

		}

		messageHandle.mu.Lock()

		if len(messageHandle.MessageQue) > 1 {
			messageHandle.MessageQue = messageHandle.MessageQue[1:]
		} else {
			messageHandle.MessageQue = []message{}
		}

		messageHandle.mu.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func (ch *frontendClienthandle) recieveBidResponseFromReplicasAndSendToClient() {
	for {
		msg, err := ch.streamBidOut.Recv()
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

// REPLICA - RESULT
func forwardQueryToReplica(ch frontendClienthandle) {
	/* for {
		for {
			messageHandle.mu.Lock()

			if len(messageHandle.MessageQue) == 0 {
				messageHandle.mu.Unlock()
				break
			}

			idFromClient := messageHandle.MessageQue[0].Id
			amountFromClient := messageHandle.MessageQue[0].Amount

			messageHandle.mu.Unlock()

			// Send data over the gRPC stream to the client
			if err := ch.streamBidOut.Send(&protos.BidRequest{
				ClientId: idFromClient,
				Amount:   amountFromClient,
			}); err != nil {
				logger.ErrorLogger.Output(2, (fmt.Sprintf("Failed to send data to client: %v", err)))
				// In case of error the client would re-subscribe so close the subscriber stream
			}
			logger.InfoLogger.Println("Forwarding message to replicas..")

		}

		messageHandle.mu.Lock()

		if len(messageHandle.MessageQue) > 1 {
			messageHandle.MessageQue = messageHandle.MessageQue[1:]
		} else {
			messageHandle.MessageQue = []message{}
		}

		messageHandle.mu.Unlock()
		time.Sleep(1 * time.Second)
	} */
}

func (ch *frontendClienthandle) recieveQueryResponseFromReplicaAndSendToClient() {
	for {
		if !connected { // To avoid sending before connected.
			time.Sleep(1 * time.Second)
		} else {
			response, err := ch.streamResultOut.Recv()
			if err != nil {
				logger.ErrorLogger.Printf("Failed to receive message: %v", err)
			} else {
				Output(fmt.Sprintf("Current highest bid: %v from clientID: %v", response.HighestBid, response.HighestBidderID))
				logger.InfoLogger.Println("Succesfully recieved response from query")
			}
		}
	}
}

// SETUP
func (client *frontendClientReplica) setupBidStream() frontendClienthandle {
	streamOut, err := client.clientService.Bid(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to call AuctionhouseService: %v", err)
	}

	return frontendClienthandle{streamBidOut: streamOut}
}

func (frontendClient *frontendClientReplica) setupResultStream() frontendClienthandle {
	streamOut, err := frontendClient.clientService.Result(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to call AuctionhouseService: %v", err)
	}

	return frontendClienthandle{streamResultOut: streamOut}
}

func setupFrontendConnectionToReplica(port string) *frontendClientReplica {
	setupFClientReplicaID()

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

// CONNECTION
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
	return grpc.Dial(fmt.Sprintf(":%v", port), []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}...)
}

func connectToClient(port string) {
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

// EXTENSIONS
func addToMessageQueue(id, amount int32) {
	messageHandle.mu.Lock()

	messageHandle.MessageQue = append(messageHandle.MessageQue, message{
		Id:     id,
		Amount: amount,
	})

	logger.InfoLogger.Printf("Message successfully recieved and queued: %v\n", id)

	messageHandle.mu.Unlock()
}

func Output(input string) {
	fmt.Println(input)
}
