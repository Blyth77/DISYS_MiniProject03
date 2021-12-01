package frontend

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"google.golang.org/grpc"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"
	main "github.com/Blyth77/DISYS_MiniProject03/client"

)

var (
	numberOfReplicas     int
	bidToReplicaQueue    = make(chan *protos.BidRequest, 10)
	ResultToReplicaQueue = make(chan *protos.QueryResult, 10)
)

type Server struct {
	protos.UnimplementedAuctionhouseServiceServer
	subscribers     map[int]frontendClienthandle
	channelsHandler *main.FrontendConnection
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
		//go s.recieveQueryResponseFromReplicaAndSendToClient()
		go s.forwardBidToReplica()
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
		request := <-s.channelsHandler.bidSendChannel
		addToBidQueue(request)
	}
}

/*
// CLIENT - RESULT
func (s *Server) sendResultResponseToClient() {
	for {
		request := <-s.channelsHandler.querySendChannel
		addToResultQueue(request)
	}
}
*/
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
		bidFromReplicasStatus := make(map[protos.Status]int)

		for _, element := range s.subscribers {

			bid := element.recieveBidReplicas()
			logger.InfoLogger.Printf("Recieved BidStatusResponse from replicas. Status: %v", bid.Status)
			if bid.Status == protos.Status_NOW_HIGHEST_BIDDER {
				count := bidFromReplicasStatus[bid.Status]
				bidFromReplicasStatus[bid.Status] = count + 1
			}
			if bid.Status == protos.Status_TOO_LOW_BID {
				count := bidFromReplicasStatus[bid.Status]
				bidFromReplicasStatus[bid.Status] = count + 1
			}
		}

		var highest int
		var stat protos.Status
		for key, value := range bidFromReplicasStatus {
			if value > highest {
				highest = value
				stat = key
			}
		}

		s.channelsHandler.bidRecieveChannel <- &protos.StatusOfBid{
			Status: stat,
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

/*
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
	}
}
*/
/* func (ch *frontendClienthandle) recieveQueryResponseFromReplicaAndSendToClient() {
	for {
		// channel blocker
		/* if !connected { // To avoid sending before connected.
			sleep()
		} else {
			response, err := ch.streamResultOut.Recv()
			if err != nil {
				logger.ErrorLogger.Printf("Failed to receive message: %v", err)
			} else {
				output(fmt.Sprintf("Current highest bid: %v from clientID: %v", response.HighestBid, response.HighestBidderID))
				logger.InfoLogger.Println("Succesfully recieved response from query")
			}
		}
		sleep()
	}
}
*/

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

/*
func addToResultQueue(resultReq *protos.QueryResult) {
	ResultToReplicaQueue <- resultReq
	logger.InfoLogger.Printf("Message successfully recieved and queued for client%v\n", resultReq.ClientId)
} */
