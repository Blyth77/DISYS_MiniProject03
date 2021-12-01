package frontend

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"google.golang.org/grpc"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"

)

var (
	numberOfReplicas     int
	bidToReplicaQueue    = make(chan *protos.BidRequest, 10)
	ResultToReplicaQueue = make(chan *protos.QueryResult, 10)
)

type FrontendConnection struct {
	BidSendChannel    chan *protos.BidRequest
	BidRecieveChannel chan *protos.StatusOfBid
	QuerySendChannel     chan *protos.QueryResult
	ResultRecieveChannel chan *protos.ResponseToQuery
}

type Server struct {
	protos.UnimplementedAuctionhouseServiceServer
	subscribers     map[int]frontendClienthandle
	channelsHandler *FrontendConnection
}

type frontendClientReplica struct {
	clientService protos.AuctionhouseServiceClient
	conn          *grpc.ClientConn
}

type frontendClienthandle struct {
	streamBidOut    protos.AuctionhouseService_BidClient
	streamResultOut protos.AuctionhouseService_ResultClient
}

func StartFrontend(id int32, port string, client *FrontendConnection) {

	file, _ := os.Open("replicamanager/portlist/listOfReplicaPorts.txt")
	defer file.Close()

	scanner := bufio.NewScanner(file)

	index := 1
	logger.InfoLogger.Println("Connecting to replicas.")
	s := &Server{}
	s.channelsHandler = client
	s.subscribers = make(map[int]frontendClienthandle)

	for scanner.Scan() {

		po := scanner.Text()
		logger.InfoLogger.Printf("Trying to connect to replica on port: %v", po)

		frontendClientForReplica := setupFrontendConnectionToReplica(po)

		frontendClienthandler := frontendClienthandle{
			streamBidOut:    frontendClientForReplica.setupBidStream(),
			streamResultOut: frontendClientForReplica.setupResultStream(),
		}

		s.subscribers[index] = frontendClienthandler

		go s.recieveBidRequestFromClient()
		go s.recieveBidResponseFromReplicasAndSendToClient()
		go s.forwardBidToReplica()

		go s.recieveResultResponseToClient()
		go s.recieveQueryResponseFromReplicaAndSendToClient()
		go s.forwardQueryToReplica()

		index++
	}

	numberOfReplicas = index - 1
	logger.InfoLogger.Printf("%v replicas succesfully connected.", numberOfReplicas)

	bl := make(chan bool)
	<-bl
}

// CLIENT - BID
func (s *Server) recieveBidRequestFromClient() {
	for {
		request := <-s.channelsHandler.BidSendChannel
		addToBidQueue(request)
	}
}


// CLIENT - RESULT
func (s *Server) recieveResultResponseToClient() {
	for {
		request := <-s.channelsHandler.QuerySendChannel
		addToResultQueue(request)
	}
}

// REPLICA - BID
func (s *Server) forwardBidToReplica() {
	for {
		message := <-bidToReplicaQueue

		for key, element := range s.subscribers {
			element.streamBidOut.Send(message)
			logger.InfoLogger.Printf("Forwarding message to replicas %d", key)
		}
	}
}

func (s *Server) recieveBidResponseFromReplicasAndSendToClient() {
	for {
		// Missing checking for majority
		var bid *protos.StatusOfBid

		for _, element := range s.subscribers {
			bid = element.recieveBidReplicas()
		}

		s.channelsHandler.BidRecieveChannel <- &protos.StatusOfBid{
			Status: bid.Status,
		}
	}
}

func (cl *frontendClienthandle) recieveBidReplicas() *protos.StatusOfBid {
	msg, err := cl.streamBidOut.Recv()
	if err != nil {
		logger.ErrorLogger.Printf("Error in receiving message from server: %v", msg)
	} else {
		return msg
	}
	return nil
}


// REPLICA - RESULT
func (s *Server) forwardQueryToReplica() {
	for {
		message := <-ResultToReplicaQueue

		for key, element := range s.subscribers {
			element.streamResultOut.Send(message)
			logger.InfoLogger.Printf("Forwarding message to replicas %d", key)
		}
	}
}

func (s *Server) recieveQueryResponseFromReplicaAndSendToClient() {
	for {
		// Missing checking for majority
		var result *protos.ResponseToQuery

		for _, element := range s.subscribers {
			result = element.recieveResultReplicas()
		}


		s.channelsHandler.ResultRecieveChannel <- result
	}
}

func (cl *frontendClienthandle) recieveResultReplicas() *protos.ResponseToQuery {
	msg, err := cl.streamResultOut.Recv()
	if err != nil {
		logger.ErrorLogger.Printf("Error in receiving message from server: %v", msg)
	} else {
		return msg
	}
	return nil
}


// SETUP
func (client *frontendClientReplica) setupBidStream() protos.AuctionhouseService_BidClient {
	streamOut, err := client.clientService.Bid(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to call AuctionhouseService: %v", err)
	}
	return streamOut
}

func (frontendClient *frontendClientReplica) setupResultStream() protos.AuctionhouseService_ResultClient {
	streamOut, err := frontendClient.clientService.Result(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to call AuctionhouseService: %v", err)
	}

	return streamOut
}

func setupFrontendConnectionToReplica(port string) *frontendClientReplica {
	frontendClient, err := makeFrontendClient(port)

	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to make Client: %v", err)
	}

	return frontendClient
}

// CONNECTION
func makeFrontendClient(port string) (*frontendClientReplica, error) {
	conn, err := makeConnection(port)
	if err != nil {
		return nil, err
	}
	logger.InfoLogger.Printf("Has found connection to replica on port: %v\n", port)
	return &frontendClientReplica{
		clientService: protos.NewAuctionhouseServiceClient(conn),
		conn:          conn,
	}, nil
}

func makeConnection(port string) (*grpc.ClientConn, error) {
	return grpc.Dial(fmt.Sprintf(":%v", port), []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}...)
}

// EXTENSIONS
func addToBidQueue(bidReq *protos.BidRequest) {
	bidToReplicaQueue <- bidReq
	logger.InfoLogger.Printf("Message successfully recieved and queued for client%v\n", bidReq.ClientId)
}

func addToResultQueue(resultReq *protos.QueryResult) {
	ResultToReplicaQueue <- resultReq
	logger.InfoLogger.Printf("Message successfully recieved and queued for client%v\n", resultReq.ClientId)
} 
