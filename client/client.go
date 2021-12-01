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
	ID int32
)

func main() {
	port := fmt.Sprintf(":%v", os.Args[1])

	welcomeMsg()
	setup()
	client := &frontend.FrontendConnection{}

	client.BidSendChannel = make(chan *protos.BidRequest)
	client.BidRecieveChannel = make(chan *protos.StatusOfBid)
	client.QuerySendChannel = make(chan *protos.QueryResult)
	client.ResultRecieveChannel = make(chan *protos.ResponseToQuery)

	go frontend.StartFrontend(ID, port, client)

	logger.InfoLogger.Println(fmt.Sprintf("Client's assigned port: %v", port))

	go userInput(client)
	go receiveResultResponseFromFrontEnd(client)
	go recieveBidStatusFromFrontEnd(client)

	logger.InfoLogger.Println("Client setup completed")
	output(fmt.Sprintf("Client: %v is ready for bidding", ID))

	bl := make(chan bool)
	<-bl
}

// BID RPC
func sendBidRequestToFrontEnd(ch *frontend.FrontendConnection, amountValue int32) {
	clientMessageBox := &protos.BidRequest{ClientId: ID, Amount: amountValue}

	ch.BidSendChannel <- clientMessageBox
	logger.InfoLogger.Printf("Client id: %v has bidded %v on item", ID, amountValue)

}

func recieveBidStatusFromFrontEnd(ch *frontend.FrontendConnection) {
	for {
		bidStatus := <-ch.BidRecieveChannel

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
func sendQueryRequestResultToFrontEnd(ch *frontend.FrontendConnection) {
	queryResult := &protos.QueryResult{ClientId: ID}

	logger.InfoLogger.Printf("Sending query from client %d", ID)

	ch.QuerySendChannel <- queryResult

	logger.InfoLogger.Printf("Sending query from client %d was a succes!", ID)
}

func receiveResultResponseFromFrontEnd(ch *frontend.FrontendConnection) {
	for {
		response := <-ch.ResultRecieveChannel

		output(fmt.Sprintf("Current highest bid: %v from clientID: %v", response.HighestBid, response.HighestBidderID))
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
func userInput(ch *frontend.FrontendConnection) {
	for {
		var option string
		var amount int32

		fmt.Scanf("%s %d", &option, &amount)
		option = strings.ToLower(option)
		if option != "" || amount != 0 {
			switch {
			case option == "query":
				sendQueryRequestResultToFrontEnd(ch)
			case option == "bid":
				sendBidRequestToFrontEnd(ch, amount)
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

func quit() {
	output("Connection to server closed. Press any key to exit.\n")
	logger.InfoLogger.Printf("Client%v is quitting the Auctionhouse", ID)

	// missing safe close of connections in frontend

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
