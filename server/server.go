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

type message struct {
	auctionStatusMessage string
	highestBid           int32
	highestBidderID      int32
	item                 string
}

// Holds msg queue
type raw struct {
	MessageQue []message
	mu         sync.Mutex
}

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

var messageHandle = raw{}

// Bid is called upon a server struct and takes a AHService_Bidserver bc [....] .
// The server struct stores the highest bid.
// Client sends a bid msg (id, amount).
// first call to bid is to register - other calls places bid higher than the previous.
// check if the
func (s *Server) Bid(stream protos.AuctionhouseService_BidServer) error {

	fin := make(chan bool)
	// needs a go routine
	var bid, err = stream.Recv()
	if err != nil {
		return err
	}

	go s.HandleNewBidForClient(fin, bid, stream)
	//go s.SendBidStatusToClient(stream)

	//s.auctioneer.Store(request.ClientId, sub{stream: stream, finished: fin, id: request.ClientId})
	//addToMessageQueue(request.ClientId, 1, request.UserName, "")

	bl := make(chan error)
	return <-bl
}

// cursed?
func (s *Server) HandleNewBidForClient(fin chan (bool), bid *protos.BidRequest, srv protos.AuctionhouseService_BidServer) {
	//check if client is subscribed
	_, ok := s.auctioneer.Load(bid.ClientId)
	if !ok {
		s.auctioneer.Store(bid.ClientId, sub{streamBid: srv, finished: fin})
		logger.InfoLogger.Printf("Storing new client %v, in server map", bid.ClientId)
	}

	//Handle new bid - is bid higher than the last highest bid?
	if bid.Amount > currentHighestBidder.HighestBidAmount {
		hb1 := HighestBidder{
			HighestBidAmount: bid.Amount,
			HighestBidderID:  bid.ClientId,
			streamBid:        srv,
		}
		currentHighestBidder = hb1
	}

	logger.InfoLogger.Printf("Storing new bid %d for client %d in server map", bid.Amount, bid.ClientId)

	s.SendBidStatusToClient(srv, bid.ClientId)
	}

func (s *Server) SendBidStatusToClient(stream protos.AuctionhouseService_BidServer, currentBidderID int32) {
	for {

		// check for status:
		// NOW_HIGHEST_BIDDER, TOO_LOW_BID, EXCEPTION
		var status protos.Status

		switch {
		case currentHighestBidder.HighestBidderID == currentBidderID:
			status = protos.Status_NOW_HIGHEST_BIDDER

		case currentHighestBidder.HighestBidderID != currentBidderID:
			status = protos.Status_TOO_LOW_BID
		default:
			// case EXCEPTION
		}

		bidStatus := &protos.StatusOfBid{
			Status: status,
		}

		stream.Send(bidStatus)
	}
}

/* // To be used in results.
// Sends(/Bodcast) msg to all clients - is not needed anyways? see todo.txt
func (s *Server) sendResultToAll(srv protos.AuctionhouseService_ResultServer) {
	for {
		for {

			time.Sleep(500 * time.Millisecond)

			messageHandle.mu.Lock()

			// To be done in results
			// Check if there is any messages to broadcast
			if len(messageHandle.MessageQue) == 0 {
				messageHandle.mu.Unlock()
				break
			}

			auctionStatusMessage := messageHandle.MessageQue[0].auctionStatusMessage
			highestBid := messageHandle.MessageQue[0].highestBid
			highestBidderID := messageHandle.MessageQue[0].highestBidderID
			item := messageHandle.MessageQue[0].item

			messageHandle.mu.Unlock()

			s.auctioneer.Range(func(k, v interface{}) bool {
				id, ok := k.(int32)
				if !ok {
					logger.ErrorLogger.Println(fmt.Sprintf("Failed to cast auctioneer id: %T", k))
					return false
				}
				sub, ok := v.(sub)
				if !ok {
					logger.ErrorLogger.Println(fmt.Sprintf("Failed to cast auctioneer id: %T to sub", v))
					return false
				}
				// Send data over the gRPC stream to the client
				if err := sub.streamResult.Send(&protos.ResponseToQuery{
					AuctionStatusMessage: auctionStatusMessage,
					HighestBid:           highestBid,
					HighestBidderID:      highestBidderID,
					Item:                 item,
				}); err != nil {
					s.unsubscribe = append(s.unsubscribe, id)
					logger.ErrorLogger.Output(2, (fmt.Sprintf("Failed to send data to client: %v", err)))
				}
				return true
			})

			// Deletes just broadcasted message
			messageHandle.mu.Lock()

			if len(messageHandle.MessageQue) > 1 {
				messageHandle.MessageQue = messageHandle.MessageQue[1:] // delete the message at index 0 after sending to receiver
			} else {
				messageHandle.MessageQue = []message{}
			}

			messageHandle.mu.Unlock()
		}
		time.Sleep(100 * time.Millisecond)
	}
} */

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
			addToMessageQueue(0, id, "client", "has been killed") // ændres
		}
		s.auctioneer.Delete(id)
		logger.InfoLogger.Printf("Client id %v has been killed and deleted", id)
	}
}

// When time has runned out : brodcast who the winner is
func (s *Server) ResultsWhenTimeRunsOut(stream protos.AuctionhouseService_ResultServer) error {
	er := make(chan error)

	go s.receiveQueryForResultAndSendToClient(stream, er)
	//go sendToStream(stream, er) // back to bid
	//go s.sendResultToAll(stream) // send to all when the auction is done

	return <-er
}

// wait for a client to ask for the highest bidder and sends the result back
func (s *Server) receiveQueryForResultAndSendToClient(srv protos.AuctionhouseService_ResultServer, er_ chan error) {
	for {
		_, err := srv.Recv()
		if err != nil {
			break
		}
		queryResponse := &protos.ResponseToQuery{
			AuctionStatusMessage: "",
			HighestBid:           currentHighestBidder.HighestBidAmount,
			HighestBidderID:      currentHighestBidder.HighestBidderID,
			Item:                 "",
		}
		srv.Send(queryResponse)
	}
}

//Add a msg to a queue for processing all messages
func addToMessageQueue(highestBid, highestBidderID int32, auctionStatusMessage, item string) {
	messageHandle.mu.Lock()

	messageHandle.MessageQue = append(messageHandle.MessageQue, message{
		auctionStatusMessage: auctionStatusMessage,
		highestBid:           highestBid,
		highestBidderID:      highestBidderID,
		item:                 item,
	})

	// logger.InfoLogger.Printf("Message successfully recieved and queued: %v\n", id)

	messageHandle.mu.Unlock()
}

// outcommented if parts are needed later on
/* func sendToStream(srv protos.AuctionhouseService_ResultServer, er_ chan error) {
	for {
		time.Sleep(500 * time.Millisecond)

		err := srv.Send(&protos.StatusMessage{
			Operation: "Publish()",
			Status:    protos.Status_SUCCESS,
		})

		if err != nil {
			logger.InfoLogger.Println(fmt.Sprintf("An error occured when sending message: %v", err))
			er_ <- err
		}
	}
}
*/
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

	var o string
	fmt.Scanln(&o)
	os.Exit(3)
}

func Output(input string) {
	fmt.Println(input)
}
