package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	//google_protobuf "github.com/golang/protobuf/ptypes/empty"
)

var (
	port       = ":3000"
	clientName string
	ID         int32
)

type AuctionClient struct {
	clientService protos.AuctionhouseServiceClient
	conn          *grpc.ClientConn
}

type clienthandle struct {
	streamBidOut    protos.AuctionhouseService_BidClient
	streamResultOut protos.AuctionhouseService_ResultClient
}

func main() {
	Output(WelcomeMsg())

	setup()
	ID = int32(rand.Intn(1e4))

	logger.LogFileInit("client", ID)

	client, err := makeClient()
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to make Client: %v", err)
	}

	client.EnterUsername()

	channelBid := client.setupBidStream()
	channelResult := client.setupResultStream()

	//Query result
	go channelResult.sendQueryResult(*client)
	go channelResult.receiveFromResult()

	// BID
	go channelBid.sendMessageBid(*client)
	go channelBid.recvBidStatus()

	bl := make(chan bool)
	<-bl
}

func (client *AuctionClient) setupBidStream() clienthandle {
	streamOut, err := client.clientService.Bid(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to call AuctionhouseService: %v", err)
	}

	ch := clienthandle{
		streamBidOut: streamOut,
	}
	return ch
}

func (client *AuctionClient) setupResultStream() clienthandle {
	streamOut, err := client.clientService.Result(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to call AuctionhouseService: %v", err)
	}

	ch := clienthandle{
		streamResultOut: streamOut,
	}
	return ch
}

// Ask server (by sending query msg w. client id) to send result msg (includes: auctionStatusMessage, highest bid,
// id of the client w. the highest bid and the item for which they are bidding on)
func (ch *clienthandle) sendQueryResult(client AuctionClient) {
	for {

		queryResult := &protos.QueryResult{
			ClientId: ID,
		}

		err := ch.streamResultOut.Send(queryResult)
		if err != nil {
			logger.ErrorLogger.Printf("Error while sending result query message to server :: %v", err)
		}
	}
}

//send result msg, when queried by client or time for item has runned out
// TODO : IMPLEMENT
func (ch *clienthandle) receiveFromResult() {
	for {
		response, err := ch.streamResultOut.Recv()
		if err != nil {
			logger.WarningLogger.Printf("Failed to receive message: %v", err)
		}

		Output(fmt.Sprintf("Highest bid: %v", response.HighestBid)) // selvføli det ska.. der ska mere her ik

		/* 	string auctionStatusMessage = 1;
		int32 highestBid = 2;
		int32 highestBidderID = 3;
		string item = 4; */
	}
}

// Client send bid request incl. userinput: amount
func (ch *clienthandle) sendMessageBid(client AuctionClient) {
	for {
		amount, _ := strconv.Atoi(UserInput())

		clientMessageBox := &protos.BidRequest{
			ClientId: ID,
			Amount:   int32(amount),
		}

		err := ch.streamBidOut.Send(clientMessageBox)
		if err != nil {
			Output("An error occured while bidding, please try again")
			logger.WarningLogger.Printf("Error while sending message to server: %v", err)
		} else {
			logger.InfoLogger.Printf("Client id: %v has bidded %v in currency on item", ID, amount)
		}
	}
}

// When client has sent a bid request - recieves a status message: success, fail or expection
func (ch *clienthandle) recvBidStatus() {
	for {
		msg, err := ch.streamBidOut.Recv()
		if err != nil {
			logger.InfoLogger.Printf("Error in receiving message from server: %v", msg)
			Output("Server recieved bid!") //Maybe says more things!
		}
	}
}

//Connects and creates client through protos.NewAuctionhouseServiceClient(connection)
func makeClient() (*AuctionClient, error) {
	conn, err := makeConnection()
	if err != nil {
		return nil, err
	}
	return &AuctionClient{
		clientService: protos.NewAuctionhouseServiceClient(conn),
		conn:          conn,
	}, nil
}

func makeConnection() (*grpc.ClientConn, error) {
	logger.InfoLogger.Print("Connecting to the auctionhouse...")
	return grpc.Dial(port, []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}...)
}

func WelcomeMsg() string {
	return `
______________________________________________________
======================================================
    **>>> WELCOME TO BARBETTES AUCTIONHOUSE <<<**
======================================================
Please enter an username to begin:`
}

func (s *AuctionClient) EnterUsername() {
	clientName = UserInput()
	Welcome(clientName)
	logger.InfoLogger.Printf("User registred: %s", clientName)
	println(clientName)
}

func UserInput() string {
	reader := bufio.NewReader(os.Stdin)
	msg, err := reader.ReadString('\n')
	if err != nil {
		logger.ErrorLogger.Printf(" Failed to read from console: %v", err)
	}
	msg = strings.Trim(msg, "\r\n")

	return msg
}

func Welcome(input string) {
	Output("Type: '-- quit' to exit")
}

func FormatToChat(user, msg string, timestamp int32) string {
	return fmt.Sprintf("%d - %v:  %v", timestamp, user, msg)
}

func Output(input string) {
	fmt.Println(input)
}

func setup() {
	rand.Seed(time.Now().UnixNano())
}
