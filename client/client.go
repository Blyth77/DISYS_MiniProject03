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

type clientStruct struct {
	ID             int32
	auctionRunning bool
}

func main() {
	port := fmt.Sprintf(":%v", os.Args[1])

	welcomeMsg()

	clientStruct := clientStruct{}
	clientStruct.setup()

	client := &frontend.FrontendConnection{}
	setupClientChannels(client)

	go frontend.StartFrontend(clientStruct.ID, port, client)
	logger.InfoLogger.Printf("Client's assigned port: %v\n", port)

	go clientStruct.userInput(client)
	go clientStruct.receiveResultResponseFromFrontEnd(client)
	go clientStruct.recieveBidStatusFromFrontEnd(client)

	logger.InfoLogger.Println("Client setup completed")
	output(fmt.Sprintf("Client: %v is ready for bidding", clientStruct.ID))

	bl := make(chan bool)
	<-bl
}

// BID RPC
func (c *clientStruct) sendBidRequestToFrontEnd(ch *frontend.FrontendConnection, amountValue int32) {
	message := &protos.BidRequest{ClientId: c.ID, Amount: amountValue}

	ch.BidSendChannel <- message
	logger.InfoLogger.Printf("Client id: %v has bidded %v on item\n", c.ID, amountValue)
}

func (c *clientStruct) recieveBidStatusFromFrontEnd(ch *frontend.FrontendConnection) {
	for {
		bidStatus := <-ch.BidRecieveChannel

		switch bidStatus.Status {
		case protos.Status_NOW_HIGHEST_BIDDER:
			output("We have recieved your bid! You now the highest bidder!")
		case protos.Status_TOO_LOW_BID:
			output("We have recieved your bid! Your bid was to low")
		case protos.Status_EXCEPTION:
			output("Something went wrong, bid not accepted by the auctionhouse")
		case protos.Status_FINISHED:
			output("No current auction. Try again later.")
			c.auctionRunning = false
			continue
		}
		logger.InfoLogger.Printf("Client%v recieved BidStatusResponse from frontend. Status: %v\n", c.ID, bidStatus.Status)
		c.auctionRunning = true
	}
}

// RESULT RPC
func (c *clientStruct) sendQueryRequestResultToFrontEnd(ch *frontend.FrontendConnection) {
	queryResult := &protos.QueryResult{ClientId: c.ID}
	logger.InfoLogger.Printf("Sending query from client %d\n", c.ID)

	ch.QuerySendChannel <- queryResult
	logger.InfoLogger.Printf("Sending query from client %d was a succes!\n", c.ID)
}

func (c *clientStruct) receiveResultResponseFromFrontEnd(ch *frontend.FrontendConnection) {
	for {
		response := <-ch.ResultRecieveChannel
		if c.auctionRunning {
			if response.AuctionStatusMessage == "finished" {
				output(fmt.Sprintf("Auction finished. ClientID: %v won %v with highest bid: %v", response.HighestBidderID, response.Item, response.HighestBid))
				logger.InfoLogger.Printf("Auction finished")
				c.auctionRunning = false
			} else {
				output(fmt.Sprintf("Current highest bid: %v from clientID: %v on %v. Time left: %v", response.HighestBid, response.HighestBidderID, response.Item, response.AuctionStatusMessage))
				logger.InfoLogger.Printf("Client%v succesfully recieved QueryResponse from frontend\n", c.ID)
				c.auctionRunning = true
			}
		} else {
			if response.AuctionStatusMessage == "new" {
				output(fmt.Sprintf("New Auction. Item: %v", response.Item))
				logger.InfoLogger.Printf("New auction started. Item: %v", response.Item)
				c.auctionRunning = true
			} else {
				output("No auction currently try bidding later, to see if any new items are up for auction.")
			}
		}
	}
}

// SETUP
func (c *clientStruct) setup() {
	rand.Seed(time.Now().UnixNano())
	c.ID = int32(rand.Intn(1e4))
	logger.LogFileInit("client", c.ID)
	c.auctionRunning = true
}

func setupClientChannels(f *frontend.FrontendConnection) {
	f.BidSendChannel = make(chan *protos.BidRequest)
	f.BidRecieveChannel = make(chan *protos.StatusOfBid)
	f.QuerySendChannel = make(chan *protos.QueryResult)
	f.ResultRecieveChannel = make(chan *protos.ResponseToQuery)
	f.QuitChannel = make(chan bool)
}

// EXTENSIONS
func (c *clientStruct) userInput(ch *frontend.FrontendConnection) {
	for {
		time.Sleep(2 * time.Second)
		var option string
		var amount int32

		fmt.Scanf("%s %d", &option, &amount)
		option = strings.ToLower(option)
		if option != "" || amount != 0 {
			switch {
			case option == "query":
				c.sendQueryRequestResultToFrontEnd(ch)
			case option == "bid":
				c.sendBidRequestToFrontEnd(ch, amount)
			case option == "quit":
				c.quit(ch)
			case option == "help":
				help()
			default:
				logger.InfoLogger.Printf("Client%v failed to parse userinput", c.ID)
				output("Did not understand, pleasy try again. Type \"help\" for help.")
			}
		}
	}
}

func (c *clientStruct) quit(ch *frontend.FrontendConnection) {
	ch.QuitChannel <- false
	output("Connection to server closed. Press any key to exit.\n")
	logger.InfoLogger.Printf("Client%v is quitting the Auctionhouse", c.ID)

	ch.QuerySendChannel <- nil
	ch.BidSendChannel <- nil

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
