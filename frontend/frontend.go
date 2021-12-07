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

type FrontendConnection struct {
	BidSendChannel       chan *protos.BidRequest
	BidRecieveChannel    chan *protos.StatusOfBid
	QuerySendChannel     chan *protos.QueryResult
	ResultRecieveChannel chan *protos.ResponseToQuery
	QuitChannel          chan bool
}

type Server struct {
	protos.UnimplementedAuctionhouseServiceServer
	replicas             map[int]frontendClienthandle
	channelsHandler      *FrontendConnection
	bidMajority          map[protos.Status]int
	bidToReplicaQueue    chan *protos.BidRequest
	resultToReplicaQueue chan *protos.QueryResult
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
	logger.InfoLogger.Println("Connecting to replicas.")
	s := &Server{}
	s.setupServer(client)

	go s.connectingToServer()
	go s.recieveBidRequestFromClient()
	go s.forwardBidToReplica()
	go s.recieveResultResponseAndSendToClient()
	go s.forwardQueryToReplica()

	go client.closing(s)
	<-client.QuitChannel
	logger.InfoLogger.Println("Closing frontend.")
}

// CLIENT - BID
func (s *Server) recieveBidRequestFromClient() {
	for {
		request := <-s.channelsHandler.BidSendChannel
		if request == nil {
			break
		}
		s.addToBidQueue(request)
	}
}

// CLIENT - RESULT
func (s *Server) recieveResultResponseAndSendToClient() {
	for {
		request := <-s.channelsHandler.QuerySendChannel
		if request == nil {
			break
		}
		s.addToResultQueue(request)
	}
}

// REPLICA - BID
func (s *Server) forwardBidToReplica() {
	for {
		message := <-s.bidToReplicaQueue

		for key, element := range s.replicas {
			element.streamBidOut.Send(message)
			logger.InfoLogger.Printf("Forwarding message to replicas%d. CliendID: %v Amount: %d  ", key, message.ClientId, message.Amount)
		}
		s.recieveBidResponseFromReplicasAndSendToClient()
	}
}

func (s *Server) recieveBidResponseFromReplicasAndSendToClient() {
	for key, element := range s.replicas {
		bid, err := element.recieveBidReplicas()
		if err != nil {
			element.streamResultOut.CloseSend()
			element.streamBidOut.CloseSend()

			delete(s.replicas, key)
			logger.InfoLogger.Printf("Removed failed replica: replica%v!", key)
			continue
		}
		s.fillMajorityList(bid.Status)
	}

	s.channelsHandler.BidRecieveChannel <- &protos.StatusOfBid{
		Status: s.checkMajority(),
	}
	s.clearMajorityValues()
}

func (cl *frontendClienthandle) recieveBidReplicas() (*protos.StatusOfBid, error) {
	msg, err := cl.streamBidOut.Recv()
	if err != nil {
		logger.ErrorLogger.Printf("Error in receiving message from server")
		return &protos.StatusOfBid{
			Status: protos.Status_EXCEPTION,
		}, err
	} else {
		logger.InfoLogger.Printf("Receiving message from server: %v", msg.Status.String())
		return msg, err
	}
}

// REPLICA - RESULT
func (s *Server) forwardQueryToReplica() {
	for {
		message := <-s.resultToReplicaQueue

		for key, element := range s.replicas {
			element.streamResultOut.Send(message)
			logger.InfoLogger.Printf("Forwarding message to replicas %d", key)
		}
		s.recieveQueryResponseFromReplicaAndSendToClient()
	}
}

func (s *Server) recieveQueryResponseFromReplicaAndSendToClient() {
	var result *protos.ResponseToQuery

	// Reds from one of the replicas. Majority is ignored, but replica is disconnected if there is an error.
	for _, element := range s.replicas {
		result = element.recieveResultReplicas()
		if result == nil {
			element.streamResultOut.CloseSend()
			element.streamBidOut.CloseSend()
		} else {
			s.channelsHandler.ResultRecieveChannel <- result
			return
		}
	}
}

func (cl *frontendClienthandle) recieveResultReplicas() *protos.ResponseToQuery {
	msg, err := cl.streamResultOut.Recv()
	if err != nil {
		logger.ErrorLogger.Printf("Error in receiving message from server")
		return nil
	} else {
		return msg
	}
}

// SETUP
func (client *frontendClientReplica) setupBidStream() (protos.AuctionhouseService_BidClient, error) {
	streamOut, err := client.clientService.Bid(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to call AuctionhouseService: %v", err)
	}
	return streamOut, err
}

func (frontendClient *frontendClientReplica) setupResultStream() (protos.AuctionhouseService_ResultClient, error) {
	streamOut, err := frontendClient.clientService.Result(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to call AuctionhouseService: %v", err)
	}

	return streamOut, err
}

func (s *Server) setupServer(client *FrontendConnection) {
	s.channelsHandler = client
	s.replicas = make(map[int]frontendClienthandle)
	s.bidMajority = make(map[protos.Status]int)
	s.bidToReplicaQueue = make(chan *protos.BidRequest, 10)
	s.resultToReplicaQueue = make(chan *protos.QueryResult, 10)
	s.setupMajority()
}

func (s *Server) setupMajority() {
	s.bidMajority[protos.Status_NOW_HIGHEST_BIDDER] = 0
	s.bidMajority[protos.Status_TOO_LOW_BID] = 0
	s.bidMajority[protos.Status_EXCEPTION] = 0
	s.bidMajority[protos.Status_FINISHED] = 0
	s.bidMajority[protos.Status_NEW] = 0

}

// CONNECTION
func (s *Server) connectingToServer() {
	file, _ := os.Open("replicamanager/portlist/listOfReplicaPorts.txt")
	defer file.Close()
	index := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		po := scanner.Text()
		logger.InfoLogger.Printf("Trying to connect to replica on port: %v", po)

		frontendClientForReplica := setupFrontendConnectionToReplica(po)

		bidStream, err := frontendClientForReplica.setupBidStream()
		if err != nil {
			continue
		}
		resultStream, err := frontendClientForReplica.setupResultStream()
		if err != nil {
			continue
		}

		frontendClienthandler := frontendClienthandle{
			streamBidOut:    bidStream,
			streamResultOut: resultStream,
		}

		s.replicas[index] = frontendClienthandler
		index++
	}
}

func setupFrontendConnectionToReplica(port string) *frontendClientReplica {
	conn, err := makeConnection(port)
	if err != nil {
		logger.InfoLogger.Printf("Failed to connect to replica on port: %v", port)
		return nil
	}
	logger.InfoLogger.Printf("Has found connection to replica on port: %v\n", port)
	return &frontendClientReplica{
		clientService: protos.NewAuctionhouseServiceClient(conn),
		conn:          conn,
	}
}

func makeConnection(port string) (*grpc.ClientConn, error) {
	return grpc.Dial(fmt.Sprintf(":%v", port), []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}...)
}

// EXTENSIONS
func (s *Server) addToBidQueue(bidReq *protos.BidRequest) {
	s.bidToReplicaQueue <- bidReq
	logger.InfoLogger.Printf("BidMessage successfully recieved and queued from client%v\n", bidReq.ClientId)
}

func (s *Server) addToResultQueue(resultReq *protos.QueryResult) {
	s.resultToReplicaQueue <- resultReq
	logger.InfoLogger.Printf("ResultMessage successfully recieved and queued from client%v\n", resultReq.ClientId)
}

func (ch *FrontendConnection) closing(s *Server) {
	<-ch.QuitChannel
	for _, conn := range s.replicas {
		conn.streamBidOut.CloseSend()
		conn.streamResultOut.CloseSend()
	}
}

func (s *Server) clearMajorityValues() {
	for k := range s.bidMajority {
		s.bidMajority[k] = 0
	}
}

func (s *Server) fillMajorityList(bid protos.Status) {
	for k := range s.bidMajority {
		if bid == k {
			s.bidMajority[k] = s.bidMajority[k] + 1
		}
	}
}

func (s *Server) checkMajority() protos.Status {
	var finalBid protos.Status
	var count int

	for k, v := range s.bidMajority {
		if v > count {
			finalBid = k
			count = v
		}
	}
	return finalBid
}

