package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"
	frontend "github.com/Blyth77/DISYS_MiniProject03/frontend"

)

var (
	ID int32
)

type FrontendConnection struct {
	bidSendChannel    chan *protos.BidRequest
	bidRecieveChannel chan *protos.StatusOfBid
	//querySendChannel     chan *protos.QueryResult
	//resultRecieveChannel chan *protos.ResponseToQuery
}

func main() {
	port := fmt.Sprintf(":%v", os.Args[1])

	welcomeMsg()
	setup()
	client := &FrontendConnection{}

	client.bidSendChannel = make(chan *protos.BidRequest)
	client.bidRecieveChannel = make(chan *protos.StatusOfBid)
	//client.querySendChannel = make(chan *protos.QueryResult)
	//client.resultRecieveChannel = make(chan *protos.ResponseToQuery)

	go frontend.StartFrontend(ID, port, client)

	logger.InfoLogger.Println(fmt.Sprintf("Client's assigned port: %v", port))

	go client.userInput()
	go client.receiveResultResponseFromFrontEnd()
	go client.recieveBidStatusFromFrontEnd()

	logger.InfoLogger.Println("Client setup completed")
	output(fmt.Sprintf("Client: %v is ready for bidding", ID))

	bl := make(chan bool)
	<-bl
}

// BID RPC
func (ch *FrontendConnection) sendBidRequestToFrontEnd(amountValue int32) {
	clientMessageBox := &protos.BidRequest{ClientId: ID, Amount: amountValue}

	ch.bidSendChannel <- clientMessageBox
	logger.InfoLogger.Printf("Client id: %v has bidded %v on item", ID, amountValue)

}

// tog clienthandle før
func (ch *FrontendConnection) recieveBidStatusFromFrontEnd() {
	for {
		bidStatus := <-ch.bidRecieveChannel

		switch bidStatus.Status {
		case protos.Status_NOW_HIGHEST_BIDDER:
			output("We have recieved your bid! You now the highest bidder!")
		case protos.Status_TOO_LOW_BID:
			output("We have recieved your bid! Your bid was to low")
		case protos.Status_EXCEPTION:
			output("Something went wrong, bid not accepted by the auctionhouse")
		}
		logger.InfoLogger.Println(fmt.Sprintf("Client%v recieved BidStatusResponse from frontend. Status: %v", ID, bidStatus.Status))
	}

}

// RESULT RPC
func (ch *FrontendConnection) sendQueryRequestResultToFrontEnd() {
	//queryResult := &protos.QueryResult{ClientId: ID}

	logger.InfoLogger.Printf("Sending query from client %d", ID)

	//ch.querySendChannel <- queryResult

	logger.InfoLogger.Printf("Sending query from client %d was a succes!", ID)
}

func (ch *FrontendConnection) receiveResultResponseFromFrontEnd() {
	for {
		//response := <-ch.resultRecieveChannel

		//output(fmt.Sprintf("Current highest bid: %v from clientID: %v", response.HighestBid, response.HighestBidderID))
		logger.InfoLogger.Println(fmt.Sprintf("Client%vSuccesfully recieved QueryResponse from frontend", ID))

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
func (ch *FrontendConnection) userInput() {
	for {
		var option string
		var amount int32

		fmt.Scanf("%s %d", &option, &amount)
		option = strings.ToLower(option)
		if option != "" || amount != 0 {
			switch {
			case option == "query":
				ch.sendQueryRequestResultToFrontEnd()
			case option == "bid":
				ch.sendBidRequestToFrontEnd(amount)
			case option == "quit":
				quit() // Cause system to fuck up!
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
func quit() {
	output("Connection to server closed. Press any key to exit.\n")
	logger.InfoLogger.Printf("Client%v is quitting the Auctionhouse", ID)

	var o string
	fmt.Scanln(&o)
	os.Exit(3)
}

func output(input string) {
	fmt.Println(input)
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
