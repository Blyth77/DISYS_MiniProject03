package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	//google_protobuf "github.com/golang/protobuf/ptypes/empty"
)

var (
	checkingStatus bool
	connected      bool
	clientName     string
	ID             int32 
)

type AuctionClient struct {
	clientService protos.AuctionhouseServiceClient
	conn          *grpc.ClientConn 
}

type clienthandle struct {
	streamOut protos.AuctionhouseService_PublishClient
}

func main() {
	logger.LogFileInit()
	Output(WelcomeMsg())

	rand.Seed(time.Now().UnixNano())
	ID = int32(rand.Intn(1e6))

	client, err := makeClient()
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to make Client: %v", err)
	}

	client.EnterUsername()

	streamOut, err := client.clientService.Publish(context.Background())
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to call AuctionhouseService :: %v", err)
	}

	ch := clienthandle{
		streamOut: streamOut,
	}

	go client.receiveMessage()
	go ch.sendMessage(*client)
	go ch.recvStatus()

	//blocker
	bl := make(chan bool)
	<-bl
}

func (cc *AuctionClient) receiveMessage() {
	var err error
	var stream protos.AuctionhouseService_BroadcastClient

	for {
		if stream == nil {
			if stream, err = cc.subscribe(); err != nil {
				UserInput() 
				Output("Closing client")
				logger.ErrorLogger.Fatalf("Failed to join: %v", err)
				cc.sleep()
				continue
			}
		}
		response, err := stream.Recv()

		if err != nil {
			logger.WarningLogger.Printf("Failed to receive message: %v", err)
			stream = nil
			cc.sleep()
			continue
		}

		msgCode := response.Code
		switch {
		case msgCode == 1:
			Output(fmt.Sprintf("%s joined the auctionhouse\n", response.Username))
		case msgCode == 2 && response.ClientId != ID:
			// Output(fmt.Sprintf(" %s says: %s \n", response.Username, response.Msg))
		case msgCode == 3:
			// Output(fmt.Sprintf("Logical Timestamp:, %s left the server\n", response.Username))
		case msgCode == 4:
			// Output("Logical Timestamp:, server closed. Press ctrl + c to exit.\n")
		}
	}
}

func (c *AuctionClient) subscribe() (protos.AuctionhouseService_BroadcastClient, error) {
	logger.InfoLogger.Printf("Buyers client ID: %d", ID)
	return c.clientService.Broadcast(context.Background(), &protos.Subscription{
		ClientId: ID,
		UserName: clientName,
	})
}

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
	return grpc.Dial(":3000", []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}...)
}

func (ch *clienthandle) recvStatus() {
	for !connected {
		mssg, err := ch.streamOut.Recv()
		if err != nil {
			logger.ErrorLogger.Fatalf("Error in receiving message from server :: %v", err)
		}
		if checkingStatus {
			Output(fmt.Sprintf("%s : %s \n", mssg.Operation, mssg.Status))
		}
		connected = true
	}
}

func (ch *clienthandle) sendMessage(client AuctionClient) {
	for {
		clientMessage := UserInput()

		if strings.Contains(clientMessage, "-- quit") {
			Output("Logical Timestamp:%d, connection to server closed. Press any key to exit.\n")
			clientMessageBox := &protos.ClientMessage{
				ClientId:         ID,
				UserName:         clientName,
				Msg:              "",
				Code:             2,
			}

			err := ch.streamOut.Send(clientMessageBox)
			if err != nil {
				logger.WarningLogger.Printf("Error while sending message to server :: %v", err)
			}
			UserInput()

			os.Exit(3)

		} else {

			clientMessageBox := &protos.ClientMessage{
				ClientId:         ID,
				UserName:         clientName,
				Msg:              clientMessage,
				Code:             1,
			}

			err := ch.streamOut.Send(clientMessageBox)
			if err != nil {
				logger.WarningLogger.Printf("Error while sending message to server :: %v", err)
			}
		}
	}
}

func (c *AuctionClient) sleep() {
	time.Sleep(time.Second * 2)
}

func WelcomeMsg() string {
	return `
______________________________________________________
======================================================
    **>>> WELCOME TO BARBETTES AUCTIONHOUSE <<<**
======================================================
Please enter an username to begin:`
}

func LimitReader(s string) string {
	limit := 128

	reader := strings.NewReader(s)

	buff := make([]byte, limit)

	n, _ := io.ReadAtLeast(reader, buff, limit)

	if n != 0 {
		return string(buff)
	} else {
		return s
	}
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
		logger.ErrorLogger.Printf(" Failed to read from console :: %v", err)
	}
	msg = strings.Trim(msg, "\r\n")

	return LimitReader(msg)
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
