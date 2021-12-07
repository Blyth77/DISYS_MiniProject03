package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"
)

type Server struct {
	protos.UnimplementedAuctionhouseServiceServer
	ID                   int32
	currentHighestBidder HighestBidder
	alive                bool
	auctionTime          int
	timeRunning          bool
	finished             bool
	item                 string
	startingNew          bool
	mu                   sync.Mutex
}

type HighestBidder struct {
	HighestBidAmount int32
	HighestBidderID  int32
	streamBid        protos.AuctionhouseService_BidServer
}

func main() {
	port, _ := strconv.Atoi(os.Args[1])
	s := &Server{}
	s.setup()
	go s.connectingToClients(port)
	logger.InfoLogger.Printf("Replica%v ready for requests on port: %v AuctionTime: %d\n", s.ID, port, s.auctionTime)
	output(fmt.Sprintf("Replica%v connected on port: %v", s.ID, port))

	s.initateAuction()

	for {
		var o string
		fmt.Scanln(&o)
		if o == "kill" {
			output("Replica dying... ")
			s.alive = false
			time.Sleep(4 * time.Second)
			output("Exit successfull")

			logger.InfoLogger.Println("Exit successfull. Server closing...")
			os.Exit(3)
		}
		if o == "new" {

			s.mu.Lock()
			if s.finished {
				s.initateAuction()
				msg := fmt.Sprintf("Initiating new auction with: Item: %v, Duration: %d", s.item, s.auctionTime)
				output(msg)
				logger.InfoLogger.Println(msg)
			} else {
				output("Ongoing auction!")
			}
			s.mu.Unlock()
		}
	}
}

func (s *Server) initateAuction() {
	output("Please enter an item and a timespan(seconds) for the auction:")
	var item string
	var time int
	fmt.Scanf("%v %d", &item, &time)
	s.item = item
	s.auctionTime = time
	s.finished = false
	s.timeRunning = false
	s.startingNew = true
	s.currentHighestBidder = HighestBidder{
		HighestBidAmount: 0,
		HighestBidderID:  0,
	}
	logger.InfoLogger.Printf("Starting auction for %v. Time: %d", item, time)
	output(fmt.Sprintf("Starting auction for %v. Time: %d", item, time))
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
		if s.alive {
			var bid, err = srv.Recv()
			if err != nil {
				logger.ErrorLogger.Printf("FATAL: failed to recive bid from frontend: %s", err)
				continue
			} else {
				if !s.timeRunning && s.startingNew {
					go s.timer()
				}

				if s.finished {
					s.SendBidStatusToClient(srv, protos.Status_FINISHED)
				} else if bid.Amount > s.currentHighestBidder.HighestBidAmount {
					highBidder := HighestBidder{
						HighestBidAmount: bid.Amount,
						HighestBidderID:  bid.ClientId,
						streamBid:        srv,
					}
					s.currentHighestBidder = highBidder
					logger.InfoLogger.Printf("Recived new bid: %d from client%d", bid.Amount, bid.ClientId)
					s.SendBidStatusToClient(srv, protos.Status_NOW_HIGHEST_BIDDER)
					s.timeRunning = true
					s.startingNew = false
				} else {
					s.SendBidStatusToClient(srv, protos.Status_TOO_LOW_BID)
				}
			}
		} else {
			s.SendBidStatusToClient(srv, protos.Status_EXCEPTION)
		}
	}
}

func (s *Server) SendBidStatusToClient(srv protos.AuctionhouseService_BidServer, status protos.Status) {
	bidStatus := &protos.StatusOfBid{
		Status: status,
	}
	srv.Send(bidStatus)
}

// RESULT
func (s *Server) Result(stream protos.AuctionhouseService_ResultServer) error {
	er := make(chan error)

	go s.receiveQueryFromFrontEndAndSendResponse(stream, er)

	return <-er
}

func (s *Server) receiveQueryFromFrontEndAndSendResponse(srv protos.AuctionhouseService_ResultServer, er_ chan error) {
	for {
		if s.alive {
			id, err := srv.Recv()
			if err != nil {
				logger.ErrorLogger.Printf("FATAL: failed to recive QueryResult from Replica: %s", err)
				break
			} else {
				logger.InfoLogger.Printf("Query recieved from frontend%d", id.ClientId)
				if !s.finished {
					er := srv.Send(
						createQueryResponse(
							fmt.Sprintf("%d", s.auctionTime),
							s.item,
							s.currentHighestBidder.HighestBidAmount,
							s.currentHighestBidder.HighestBidderID,
						))
					if er != nil {
						logger.ErrorLogger.Fatalf("FATAL: failed to send ResponseToQuery \"MSG\" to frontend: %s", err)
					}
					logger.InfoLogger.Println("Query \"MSG\" sent to frontend ")
				} else if s.startingNew {
					er := srv.Send(
						createQueryResponse(
							"new",
							s.item,
							s.currentHighestBidder.HighestBidAmount,
							s.currentHighestBidder.HighestBidderID,
						))
					if er != nil {
						logger.ErrorLogger.Fatalf("FATAL: failed to send ResponseToQuery \"NEW\" to frontend: %s", err)
					}
					logger.InfoLogger.Println("Query \"NEW\" sent to frontend ")
				} else {
					er := srv.Send(
						createQueryResponse(
							"finished",
							s.item,
							s.currentHighestBidder.HighestBidAmount,
							s.currentHighestBidder.HighestBidderID,
						))
					if er != nil {
						logger.ErrorLogger.Fatalf("FATAL: failed to send ResponseToQuery \"FINISHED\" to frontend: %s", err)
					}
					logger.InfoLogger.Println("Query \"FINISHED\" sent to frontend")
				}
			}
		}
	}
}

// SETUP
func (s *Server) setup() {
	rand.Seed(time.Now().UnixNano())
	s.ID = int32(rand.Intn(1e4))
	logger.LogFileInit("replica", s.ID)
	s.alive = true
}

// CONNECTION
func (s *Server) connectingToClients(port int) error {
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
	bl := make(chan error)
	return <-bl
}

// EXTENSIONS
func output(input string) {
	fmt.Println(input)
}

func (s *Server) timer() {
	for s.auctionTime != 0 {
		s.auctionTime = s.auctionTime - 1
		time.Sleep(1 * time.Second)
		logger.InfoLogger.Printf("Time: %v", s.auctionTime)
	}
	s.finished = true
	logger.InfoLogger.Println("Auction finished")
}

func createQueryResponse(message, item string, amount, id int32) *protos.ResponseToQuery {
	return &protos.ResponseToQuery{
		AuctionStatusMessage: message,
		HighestBid:           amount,
		HighestBidderID:      id,
		Item:                 item,
	}
}
