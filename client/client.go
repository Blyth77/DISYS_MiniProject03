package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	frontend "github.com/Blyth77/DISYS_MiniProject03/frontend"
	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
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
	port := fmt.Sprintf(":%v", os.Args[1])
	Output(WelcomeMsg())

	go frontend.Start(ID, port)

	//--------------------
	client := setupClient(port)

	channelBid := client.setupBidStream()
	channelResult := client.setupResultStream()

	Output("Current item is: ITEM, current highest bid is: HIGHEST_BID, by client: ID") //minus

	go UserInput(client, channelBid, channelResult) // minus
	go channelResult.receiveFromResultStream()
	go channelBid.recvBidStatus()
	//______________________

	bl := make(chan bool)
	<-bl
}

func UserInput(client *AuctionClient, bid clienthandle, result clienthandle) {
	for {
		var option string
		var amount int32

		fmt.Scanf("%s %d", &option, &amount)
		option = strings.ToLower(option)
		switch {
		case option == "query":
			if !connected {
				Output("Please make a bid, before querying!")
			} else {
				result.sendQueryForResult(*client)
			}
		case option == "bid":
			sendBidRequest(*client, amount, bid)
		case option == "quit":
			Quit(client) // Cause system to fuck up!
		case option == "help":
			Help()
		default:
			Output("Did not understand, pleasy try again. Type \"help\" for help.")
		}
	}
}

func (client *AuctionClient) setupBidStream() clienthandle {
	streamOut, err := client.clientService.Bid(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to call AuctionhouseService: %v", err)
	}
	return clienthandle{streamBidOut: streamOut}
}

func (client *AuctionClient) setupResultStream() clienthandle {
	streamOut, err := client.clientService.Result(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to call AuctionhouseService: %v", err)
	}

	return clienthandle{streamResultOut: streamOut}
}

func (ch *clienthandle) sendQueryForResult(client AuctionClient) {
	queryResult := &protos.QueryResult{ClientId: ID}

	logger.InfoLogger.Printf("Sending query from client %d", ID)

	err := ch.streamResultOut.Send(queryResult)
	if err != nil {
		logger.ErrorLogger.Printf("Error while sending result query message to server :: %v", err)
		Output("Something went wrong, please try again.")
	}

	logger.InfoLogger.Printf("Sending query from client %d was a succes!", ID)
}

func (ch *clienthandle) receiveFromResultStream() {
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

func sendBidRequest(client AuctionClient, amountValue int32, ch clienthandle) {
	clientMessageBox := &protos.BidRequest{ClientId: ID, Amount: amountValue}

	err := ch.streamBidOut.Send(clientMessageBox)
	if err != nil {
		Output("An error occured while bidding, please try again")
		logger.WarningLogger.Printf("Error while sending message to server: %v", err)
	} else {
		logger.InfoLogger.Printf("Client id: %v has bidded %v on item", ID, amountValue)
	}
}

// When client has sent a bid request - recieves a status message: success, fail or expection
func (ch *clienthandle) recvBidStatus() {
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
				Output(fmt.Sprintf("We have recieved your bid! You now have the highest bid: %v", msg.HighestBid))
			case protos.Status_TOO_LOW_BID:
				Output(fmt.Sprintf("We have recieved your bid! Your bid was to low. The highest bid: %v", msg.HighestBid))
			case protos.Status_EXCEPTION:
				Output("Something went wrong, bid not accepted by the auctionhouse")
			}
			connected = true
		}
	}
}

//Connects and creates client through protos.NewAuctionhouseServiceClient(connection)
func makeClient(port string) (*AuctionClient, error) {

	conn, err := makeConnection(port)
	if err != nil {
		return nil, err
	}

	return &AuctionClient{
		clientService: protos.NewAuctionhouseServiceClient(conn),
		conn:          conn,
	}, nil
}

func makeConnection(port string) (*grpc.ClientConn, error) {
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
			query
		in the terminal, followed by enter.

	Quitting:
		To quit the auction please write:
			quit
		in the terminal, followed by enter.

	Help:
		To get the input explaination again please write:
			help
		in the terminal, followed by enter.
------------------------------------------------------------------------------------------------------------------

`
}

func Quit(client *AuctionClient) {
	client.conn.Close()
	Output("Connection to server closed. Press any key to exit.\n")

	var o string
	fmt.Scanln(&o)
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

func setupClient(port string) *AuctionClient {
	setupClientID()

	logger.LogFileInit("client", ID)

	client, err := makeClient(port)
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to make Client: %v", err)
	}

	return client
}

func setupClientID() {
	rand.Seed(time.Now().UnixNano())
	ID = int32(rand.Intn(1e4))
}
