package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	//google_protobuf "github.com/golang/protobuf/ptypes/empty"
)

var (
	port      = ":3000"
	ID        int32
	connected bool
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

	channelBid := client.setupBidStream()
	channelResult := client.setupResultStream()

	Output("Current item is: ITEM, current highest bid is: HIGHEST_BID, by client: ID")

	// UserInput
	go UserInput(client, channelBid, channelResult)

	//Query result
	go channelResult.receiveFromResult()

	// BID

	go channelBid.recvBidStatus()

	bl := make(chan bool)
	<-bl
}

func UserInput(client *AuctionClient, bid clienthandle, result clienthandle) {
	for {
		var option string
		var amount int32
		fmt.Scanf("%s %d", &option, &amount)
		switch {
		case option == "query":
			result.sendQueryResult(*client)
		case option == "bid":
			bid.sendBidRequest(*client, amount)
		case option == "q":
			Quit()
		case option == "h":
			Help()
		default:
		}
	}
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
	queryResult := &protos.QueryResult{
		ClientId: ID,
	}

	err := ch.streamResultOut.Send(queryResult)
	logger.InfoLogger.Printf("Sending query from client %d", ID)
	if err != nil {
		logger.ErrorLogger.Printf("Error while sending result query message to server :: %v", err)
	}
}

//send result msg, when queried by client or time for item has runned out
// TODO : IMPLEMENT
func (ch *clienthandle) receiveFromResult() {
	for {
		if !connected {
			time.Sleep(1*time.Second)
		} else {
			response, err := ch.streamResultOut.Recv()
			if err != nil {
				logger.WarningLogger.Printf("Failed to receive message: %v", err)
			}

			Output(fmt.Sprintf("Highest bid: %v", response.HighestBid)) // selvføli det ska.. der ska mere her ik
		}
	}
}

// Client send bid request incl. userinput: amount
func (ch *clienthandle) sendBidRequest(client AuctionClient, amountValue int32) {
	clientMessageBox := &protos.BidRequest{
		ClientId: ID,
		Amount:   amountValue,
	}

	err := ch.streamBidOut.Send(clientMessageBox)
	if err != nil {
		Output("An error occured while bidding, please try again")
		logger.WarningLogger.Printf("Error while sending message to server: %v", err)
	} else {
		logger.InfoLogger.Printf("Client id: %v has bidded %v in currency on item", ID, amountValue)
	}
}

// When client has sent a bid request - recieves a status message: success, fail or expection
func (ch *clienthandle) recvBidStatus() {
	for {
		msg, err := ch.streamBidOut.Recv()
		if err != nil {
			logger.InfoLogger.Printf("Error in receiving message from server: %v", msg)
		}
		Output(fmt.Sprintf("Server recieved bid!, %v", msg.Status)) //Maybe says more things!
		connected = true
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
Here you can bid on different items.
A certain amount of time is set off for clients to bid on an item.
The time on the items are NOT displayed to the clients, so if you wanna bid do it fast.

INPUTS
----------------------------------------------------------------------------------------------------------------
	Bidding on an item: 
		To bid on an item just write the amount in the terminal, followed by enter, the bid must be a valid int.

	Information about current item:
		To ask the auctioneer what item you are bidding on and what the highest bid is please write:
			r
		in the terminal, followed by enter.

	Quitting:
		To quit the auction please write:
			q
		in the terminal, followed by enter.

	Help:
		To get the input explaination again please write:
			h
		in the terminal, followed by enter.
------------------------------------------------------------------------------------------------------------------

`
}

func Quit() {
	Output("Connection to server closed. Press any key to exit.\n")
	os.Exit(3)
}

func Help() {
	Output(`
	This is the Auction House, here you can bid on different items.
	A certain amount of time is set off for clients to bid on an item.
	The time on the items are NOT displayed to the clients, so if you wanna bid do it fast.

	INPUTS
	----------------------------------------------------------------------------------------------------------------
		Bidding on an item: 
			To bid on an item just write the amount in the terminal, followed by enter, the bid must be a valid int.

		Information about current item:
			To ask the auctioneer what item you are bidding on and what the highest bid is please write:
				r
			in the terminal, followed by enter.

		Quitting:
			To quit the auction please write:
				q
			in the terminal, followed by enter.
		`)
}

func Output(input string) {
	fmt.Println(input)
}

func setup() {
	rand.Seed(time.Now().UnixNano())
}
