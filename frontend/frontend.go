package server
// Lav ikke nyt id
// fix børnelokker
// hvorfor grr katter den kun 2,5 gange
// HVORFOR KØRER LOOPSNE SÅ MANGE GANGE!!!!
// Hvorfor hopper barbette aldrig når hun lover det???????????
// clean up 

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
	connected            bool
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

type msgqueue struct {
	MessageQue []message
	mu         sync.Mutex
}

var messageHandle = msgqueue{}

// called by client
func Start(id int32, port string) {
	go connectToNode(port) //clienten's server og den er correct

	file, _ := os.Open("replicamanager/portlist/listOfReplicaPorts.txt")

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		scanner.Scan()
		po := scanner.Text()


		// Her skal de 3 metoder ind. og 2 go rutiner
		frontendClientForReplica := setupFrontend(po)
		Output(fmt.Sprintf("Frontend connected with replica on port: %v", po))
		channelBid := frontendClientForReplica.setupBidStream()
		channelResult := frontendClientForReplica.setupResultStream()

		go channelResult.receiveFromResultStream()
		go channelBid.recvBidStatus()
		go forwardBidToReplica(channelBid)

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

// TODO ... skal modtage fra client og forwarde til replicas.
// skal ikke store auctioneers
func (s *Server) HandleNewBidForClient(fin chan (bool), srv protos.AuctionhouseService_BidServer) {
	for {
		// får client id + amount
		var bid, err = srv.Recv()
		if err != nil {
			logger.ErrorLogger.Println(fmt.Sprintf("FATAL: failed to recive bid from client: %s", err))
		} else {
			/*
				save srv.recv info in ex an TryClientBid

				send(srv.Context())
			*/

			//check if client is subscribed
			addToMessageQueue(bid.ClientId, bid.Amount)

		}
	}
}

func addToMessageQueue(id, amount int32) {
	messageHandle.mu.Lock()

	messageHandle.MessageQue = append(messageHandle.MessageQue, message{
		Id:     id,
		Amount: amount,
	})

	logger.InfoLogger.Printf("Message successfully recieved and queued: %v\n", id)

	messageHandle.mu.Unlock()

}

func RecieveBidResponseFromReplicas(srv protos.AuctionhouseService_BidServer, clientId, amount int32) {
	// wait for ackno
	// how to get server
	//SendBidStatusToClient(srv, clientId, amount)
}

func forwardBidToReplica(ch frontendClienthandle) {
	//implement a loop
	for {
		//loop through messages in MessageQue
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

//TODO
// recieve from replica
func (s *Server) SendBidStatusToClient(stream protos.AuctionhouseService_BidServer, currentBidderID int32, currentBid int32) {

	//var status protos.Status

	// Her skal den svarer clienten NÅR!!!!! den har modtaget majority af ackno fra replicas!
	/* switch {
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
	} */

	//stream.Send(bidStatus)
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
				//HighestBid:           currentHighestBidder.HighestBidAmount,
				//HighestBidderID:      currentHighestBidder.HighestBidderID,
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
func (frontendCh *frontendClienthandle) recvBidStatus() {
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
func (frontendCh *frontendClienthandle) receiveFromResultStream() {
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
	println("Grrr cat3")

	conn, err := makeConnection(port)
	if err != nil {
		return nil, err
	}
	println("Grrr cat4")

	return &frontendClientReplica{
		clientService: protos.NewAuctionhouseServiceClient(conn),
		conn:          conn,
	}, nil
}

func makeConnection(port string) (*grpc.ClientConn, error) {
	logger.InfoLogger.Print("Connecting to the auctionhouse...")
	return grpc.Dial(fmt.Sprintf(":%v", port), []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}...)
}

func Output(input string) {
	fmt.Println(input)
}
