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

	frontend "github.com/Blyth77/DISYS_MiniProject03/frontend"
	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"
)

var (
	ID        int32
	connected bool
)

type frontendConnection struct {
	bidSendChannel       chan bidMessage
	bidRecieveChannel    chan string        // statusOfBid
	queryChannel         chan string        // chan af struct? sender dog kun sit clientID
	resultRecieveChannel chan resultMessage //the result or staus of the auction
}

type bidMessage struct {
	Id     int32
	Amount int32
}
type resultMessage struct {
	auctionStatusMessage string
	highestBid           int32
	highestBidderID      int32
	item                 string
}

type clienthandle struct {
	streamBidOut    protos.AuctionhouseService_BidClient
	streamResultOut protos.AuctionhouseService_ResultClient
}

func main() {
	port := fmt.Sprintf(":%v", os.Args[1])

	welcomeMsg()
	setup()

	go frontend.Start(ID, port)

	channelBid := client.setupBidStream()
	channelResult := client.setupResultStream()

	logger.InfoLogger.Println(fmt.Sprintf("Client's assigned port: %v", port))

	go userInput(client, channelBid, channelResult)
	go channelResult.receiveResultResponseFromFrontEnd()
	go channelBid.recieveBidStatusFromFrontEnd()

	logger.InfoLogger.Println("Client setup completed")
	output(fmt.Sprintf("Client: %v is ready for bidding", ID))

	bl := make(chan bool)
	<-bl
}

// BID RPC
func sendBidRequestToFrontEnd(client AuctionClient, amountValue int32, ch clienthandle) {
	clientMessageBox := &protos.BidRequest{ClientId: ID, Amount: amountValue}

	err := ch.streamBidOut.Send(clientMessageBox)
	if err != nil {
		output("An error occured while bidding, please try again")
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
				output("We have recieved your bid! You now the highest bidder!")
			case protos.Status_TOO_LOW_BID:
				output("We have recieved your bid! Your bid was to low")
			case protos.Status_EXCEPTION:
				output("Something went wrong, bid not accepted by the auctionhouse")
			}
			connected = true
			logger.InfoLogger.Println(fmt.Sprintf("Client%v recieved BidStatusResponse from frontend. Status: %v", ID, msg.Status))
		}
	}
}

// RESULT RPC
func (ch *clienthandle) sendQueryRequestResultToFrontEnd(queryChannel frontendConnection) {
	queryResult := &protos.QueryResult{ClientId: ID}

	logger.InfoLogger.Printf("Sending query from client %d", ID)

	err := ch.streamResultOut.Send(queryResult)
	if err != nil {
		logger.ErrorLogger.Printf("Error while sending result query message to server :: %v", err)
		output("Something went wrong, please try again.")
	}

	logger.InfoLogger.Printf("Sending query from client %d was a succes!", ID)
}

func (ch *clienthandle) receiveResultResponseFromFrontEnd() {
	for {
		if !connected { // To avoid sending before connected.
			sleep()
		} else {
			response, err := ch.streamResultOut.Recv()
			if err != nil {
				logger.ErrorLogger.Printf("Failed to receive message: %v", err)
			} else {
				output(fmt.Sprintf("Current highest bid: %v from clientID: %v", response.HighestBid, response.HighestBidderID))
				logger.InfoLogger.Println(fmt.Sprintf("Client%vSuccesfully recieved QueryResponse from frontend", ID))
			}
		}
	}
}

// SETUP
func setup() {
	rand.Seed(time.Now().UnixNano())
	ID = int32(rand.Intn(1e4))
	logger.LogFileInit("client", ID)
}

// EXTENSIONS
// skal denne tage channels i stedet?
func userInput(client *AuctionClient, bid clienthandle, result clienthandle) {
	for {
		var option string
		var amount int32

		fmt.Scanf("%s %d", &option, &amount)
		option = strings.ToLower(option)
		if option != "" || amount != 0 {
			switch {
			case option == "query":
				if !connected {
					output("Please make a bid, before querying!")
				} else {
					result.sendQueryRequestResultToFrontEnd(*client)
				}
			case option == "bid":
				sendBidRequestToFrontEnd(*client, amount, bid)
			case option == "quit":
				quit(client) // Cause system to fuck up!
			case option == "help":
				help()
			default:
				logger.InfoLogger.Printf("Client%v failed to parse userinput", ID)
				output("Did not understand, pleasy try again. Type \"help\" for help.")
			}
		}
	}
}

// behøves denne?
func quit(client *AuctionClient) {
	client.conn.Close()
	output("Connection to server closed. Press any key to exit.\n")
	logger.InfoLogger.Printf("Client%v is quitting the Auctionhouse", ID)

	var o string
	fmt.Scanln(&o)
	os.Exit(3)
}

func output(input string) {
	fmt.Println(input)
}

func sleep() {
	time.Sleep(1 * time.Second)
}

// INFO
func welcomeMsg() {
	output(`
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

`)
}

func help() {
	output(`
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
