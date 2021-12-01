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
	"time"

	"google.golang.org/grpc"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"
)

var (
	ID               int32
	numberOfReplicas int
	queue =  make(chan bidMessage, 10)
	bidQueue =  make(chan *protos.StatusOfBid, 10)
)

type Server struct {
	protos.UnimplementedAuctionhouseServiceServer
	subscribers  map[int]frontendClienthandle
}

type frontendClientReplica struct {
	clientService protos.AuctionhouseServiceClient
	conn          *grpc.ClientConn
}

type frontendClienthandle struct {
	streamBidOut    protos.AuctionhouseService_BidClient
	streamResultOut protos.AuctionhouseService_ResultClient
}

type bidMessage struct {
	Id     int32
	Amount int32
}

func Start(id int32, port string) {
	go connectToClient(port) //clienten's server

	file, _ := os.Open("replicamanager/portlist/listOfReplicaPorts.txt")
	defer file.Close()

	scanner := bufio.NewScanner(file)

	index := 1
	logger.InfoLogger.Println("Connecting to replicas.")
	s := &Server{}
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

		go s.recieveBidResponseFromReplicasAndSendToClient()
		//go s.recieveQueryResponseFromReplicaAndSendToClient()
		go s.forwardBidToReplica()
		index++
	}

	numberOfReplicas = index -1
	logger.InfoLogger.Printf("%v replicas succesfully connected.", numberOfReplicas)

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

			ID = bid.ClientId

			//check if client is subscribed
			addToMessageQueue(bid.ClientId, bid.Amount)
		}
		result := <- bidQueue
		srv.Send(result) 
	}
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
func (s *Server) forwardBidToReplica() {
	for {
		message := <- queue

		for key, element := range s.subscribers {
			element.streamBidOut.Send(&protos.BidRequest{
				ClientId: message.Id,
				Amount:   message.Amount,
			})
			logger.InfoLogger.Printf("Forwarding message to replicas %d", key)
		}	
	}

}
// Skriver
func (s *Server) recieveBidResponseFromReplicasAndSendToClient() {
	for {
		bidFromReplicasStatus := make(map[protos.Status]int)

		for _, element := range s.subscribers {

			bid := element.recieveBidReplicas()
			logger.InfoLogger.Printf("Recieved BidStatusResponse from replicas. Status: %v", bid.Status)
			if(bid.Status == protos.Status_NOW_HIGHEST_BIDDER ) {
				count := bidFromReplicasStatus[bid.Status] 
				bidFromReplicasStatus[bid.Status] = count + 1
			}
			if(bid.Status == protos.Status_TOO_LOW_BID ) {
				count := bidFromReplicasStatus[bid.Status] 
				bidFromReplicasStatus[bid.Status] = count + 1
			}
		}

		var highest int
		var stat protos.Status
		for key, value := range bidFromReplicasStatus {
			if (value > highest) {
				highest = value 
				stat = key
			}
		}
		println(stat.String())

		//fmt.Printf("Front:%v\n", stat)
		bidQueue <- &protos.StatusOfBid{
			Status: stat,
		}

		// MODULO - til at tjekke majority.
		/* if (highest > (numberOfReplicas / 2)) {
			bidQueue <- &protos.StatusOfBid{
				Status: stat,
			}
		} else {
			bidQueue <- &protos.StatusOfBid{
				Status: protos.Status_EXCEPTION,
			}
		} */
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
		sleep() */
	}
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
	logger.InfoLogger.Printf("Has found connection to replica on port: %v\n", port)
	return &frontendClientReplica{
		clientService: protos.NewAuctionhouseServiceClient(conn),
		conn:          conn,
	}, nil
}

func makeConnection(port string) (*grpc.ClientConn, error) {
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
	logger.InfoLogger.Printf("Connected to client on port: %v", port)

	bl := make(chan bool)
	<-bl
}

// EXTENSIONS
func addToMessageQueue(id, amount int32) {
	queue <- bidMessage{
		Id:     id,
		Amount: amount,
	}
	logger.InfoLogger.Printf("Message successfully recieved and queued for client%v\n", id)
}


