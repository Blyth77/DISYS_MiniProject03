/* TODO:
- FRONTEND: recvBidStatus() - skal vente på majority har acknowledged og svaret, før den godtager at de har gemt bid.
- UserInput: fix quit. 
*/

package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	frontend "github.com/Blyth77/DISYS_MiniProject03/frontend"
	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"
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
	setup()

	go frontend.Start(ID, port)

	client := setupClient(port)
	channelBid := client.setupBidStream()
	channelResult := client.setupResultStream()

	logger.InfoLogger.Println(fmt.Sprintf("Client's assigned port: %v", port))

	go UserInput(client, channelBid, channelResult)
	go channelResult.receiveResultResponseFromFrontEnd()
	go channelBid.recieveBidStatusFromFrontEnd()

	logger.InfoLogger.Println("Client setup completed")
	Output(fmt.Sprintf("Client: %v is ready for bidding", ID))

	bl := make(chan bool)
	<-bl
}

// BID RPC
func sendBidRequestToFrontEnd(client AuctionClient, amountValue int32, ch clienthandle) {
	clientMessageBox := &protos.BidRequest{ClientId: ID, Amount: amountValue}

	err := ch.streamBidOut.Send(clientMessageBox)
	if err != nil {
		Output("An error occured while bidding, please try again")
		logger.WarningLogger.Printf("Error while sending message to frontend: %v", err)
	} else {
		logger.InfoLogger.Printf("Client id: %v has bidded %v on item", ID, amountValue)
	}
}

func (ch *clienthandle) recieveBidStatusFromFrontEnd() {
	for {
		msg, err := ch.streamBidOut.Recv()
		if err != nil {
			logger.ErrorLogger.Printf("Error in receiving message from server: %v", msg)
			connected = false
		} else {
			switch msg.Status {
			case protos.Status_NOW_HIGHEST_BIDDER:
				Output(fmt.Sprintf("We have recieved your bid! You now have the highest bid: %v", msg.HighestBid))
			case protos.Status_TOO_LOW_BID:
				Output(fmt.Sprintf("We have recieved your bid! Your bid was to low. The highest bid: %v", msg.HighestBid))
			case protos.Status_EXCEPTION:
				Output("Something went wrong, bid not accepted by the auctionhouse")
			}
			connected = true
			logger.InfoLogger.Println(fmt.Sprintf("Client%v recieved BidStatusResponse from frontend. Amount: %d, Status: %v", ID, msg.HighestBid, msg.Status))
		}
	}
}


// RESULT RPC
func (ch *clienthandle) sendQueryRequestResultToFrontEnd(client AuctionClient) {
	queryResult := &protos.QueryResult{ClientId: ID}

	logger.InfoLogger.Printf("Sending query from client %d", ID)

	err := ch.streamResultOut.Send(queryResult)
	if err != nil {
		logger.ErrorLogger.Printf("Error while sending result query message to server :: %v", err)
		Output("Something went wrong, please try again.")
	}

	logger.InfoLogger.Printf("Sending query from client %d was a succes!", ID)
}

func (ch *clienthandle) receiveResultResponseFromFrontEnd() {
	for {
		if !connected { // To avoid sending before connected.
			time.Sleep(1 * time.Second)
		} else {
			response, err := ch.streamResultOut.Recv()
			if err != nil {
				logger.ErrorLogger.Printf("Failed to receive message: %v", err)
			} else {
				Output(fmt.Sprintf("Current highest bid: %v from clientID: %v", response.HighestBid, response.HighestBidderID))
				logger.InfoLogger.Println(fmt.Sprintf("Client%vSuccesfully recieved QueryResponse from frontend", ID))
			}
		}
	}
}


// CONNECTION
func makeClient(port string) (*AuctionClient, error) {
	conn, err := makeConnection(port)
	if err != nil {
		logger.ErrorLogger.Fatalf("Client%v failed to makeConnection. Error: %v", ID, err)
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


// SETUP
func setup() {
	rand.Seed(time.Now().UnixNano())
	ID = int32(rand.Intn(1e4))
	logger.LogFileInit("client", ID)
}

func setupClient(port string) *AuctionClient {

	client, err := makeClient(port)
	if err != nil {
		logger.ErrorLogger.Fatalf("Client%v setupClient failed. Error: %v", ID, err)
	}
	return client
}

func (client *AuctionClient) setupBidStream() clienthandle {
	streamOut, err := client.clientService.Bid(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Client%v failed to call AuctionhouseService bidStream. Error: %v", ID, err)
	}
	return clienthandle{streamBidOut: streamOut}
}

func (client *AuctionClient) setupResultStream() clienthandle {
	streamOut, err := client.clientService.Result(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Client%v failed to call AuctionhouseService resultStream. Error: %v", ID, err)
	}
	return clienthandle{streamResultOut: streamOut}
}


// EXTENSIONS
func UserInput(client *AuctionClient, bid clienthandle, result clienthandle) {
	for {
		var option string
		var amount int32

		fmt.Scanf("%s %d", &option, &amount)
		option = strings.ToLower(option)
		if option != "" || amount != 0 {
			switch {
			case option == "query":
				if !connected {
					Output("Please make a bid, before querying!")
				} else {
					result.sendQueryRequestResultToFrontEnd(*client)
				}
			case option == "bid":
				sendBidRequestToFrontEnd(*client, amount, bid)
			case option == "quit":
				Quit(client) // Cause system to fuck up!
			case option == "help":
				Help()
			default:
				logger.InfoLogger.Printf("Client%v failed to parse userinput", ID)
				Output("Did not understand, pleasy try again. Type \"help\" for help.")
			}
		}
	}
}

func Quit(client *AuctionClient) {
	client.conn.Close()
	Output("Connection to server closed. Press any key to exit.\n")
	logger.InfoLogger.Printf("Client%v is quitting the Auctionhouse", ID)

	var o string
	fmt.Scanln(&o)
	os.Exit(3)
}

func Output(input string) {
	fmt.Println(input)
}


// INFO
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
