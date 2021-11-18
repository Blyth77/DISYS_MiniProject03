package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	logger 
	"google.golang.org/grpc"
)

type message struct {
	ClientUniqueCode int32
	ClientName       string
	Msg              string
	MessageCode      int32
	Lamport          int32
}

type raw struct {
	MessageQue []message
	mu         sync.Mutex
}

type Server struct {
	protos.UnimplementedChittyChatServiceServer
	subscribers sync.Map
	unsubscribe []int32
	lamport     protos.LamportTimestamp
}

type sub struct {
	stream   protos.ChittyChatService_BroadcastServer
	finished chan<- bool
	name     string
}

var messageHandle = raw{}

func (s *Server) Broadcast(request *protos.Subscription, stream protos.ChittyChatService_BroadcastServer) error {
	//fmt.Printf("Receiving lamport from client: %d\n", request.LamportTimestamp)
	s.lamport.RecieveIncomingLamportInt(request.LamportTimestamp) // the server is recieving a message of a new client joining
	logger.InfoLogger.Printf("Lamp.t.: %d, Received subscribe request from ID: %d", s.lamport.Timestamp, request.ClientId)
	fin := make(chan bool)

	s.subscribers.Store(request.ClientId, sub{stream: stream, finished: fin, name: request.UserName})

	// Connecting
	addToMessageQueue(request.ClientId, s.lamport.Timestamp, 1, request.UserName, "")
	Output(fmt.Sprintf("ID: %v Name: %v, joined chat at timestamp %d", request.ClientId, request.UserName, s.lamport.Timestamp))

	go s.sendToClients(stream)

	bl := make(chan error)
	return <-bl
}

func (s *Server) sendToClients(srv protos.ChittyChatService_BroadcastServer) {
	logger.InfoLogger.Println("Request send to clients")
	//implement a loop
	for {

		//loop through messages in MessageQue
		for {

			time.Sleep(500 * time.Millisecond)

			messageHandle.mu.Lock()

			if len(messageHandle.MessageQue) == 0 {
				messageHandle.mu.Unlock()
				break
			}
			fmt.Printf("Lamport in server: %d\n", s.lamport.Timestamp)
			fmt.Printf("Lamport from message: %d\n", messageHandle.MessageQue[0].Lamport)
			s.lamport.RecieveIncomingLamportInt(messageHandle.MessageQue[0].Lamport) // dette er for at checke hvilken timestamp har max også +1 til den værdi
			fmt.Printf("Lamport from server converting to message: %d", s.lamport.Timestamp)
			messageHandle.MessageQue[0].Lamport = s.lamport.Timestamp // test
			// i dette tilfælde burde den incremente serverens timestamp blive 6+1 efter 1 besked sendt af client nr.2
			senderUniqueCode := messageHandle.MessageQue[0].ClientUniqueCode
			senderName := messageHandle.MessageQue[0].ClientName
			LamportTimestamp := messageHandle.MessageQue[0].Lamport
			messageFromServer := messageHandle.MessageQue[0].Msg
			messageCode := messageHandle.MessageQue[0].MessageCode

			messageHandle.mu.Unlock()

			s.subscribers.Range(func(k, v interface{}) bool {
				id, ok := k.(int32)
				if !ok {
					logger.InfoLogger.Println(fmt.Sprintf("Failed to cast subscriber key: %T", k))
					return false
				}
				sub, ok := v.(sub)
				if !ok {
					logger.InfoLogger.Println(fmt.Sprintf("Failed to cast subscriber value: %T", v))
					return false
				}
				// Send data over the gRPC stream to the client
				if err := sub.stream.Send(&protos.ChatRoomMessages{
					Msg:              messageFromServer,
					LamportTimestamp: LamportTimestamp,
					Username:         senderName,
					ClientId:         senderUniqueCode,
					Code:             messageCode,
				}); err != nil {
					logger.ErrorLogger.Output(2, (fmt.Sprintf("Failed to send data to client: %v", err)))
					s.unsubscribe = append(s.unsubscribe, id)
				}
				return true
			})
			logger.InfoLogger.Println("Brodcasting message success.")

			messageHandle.mu.Lock()

			if len(messageHandle.MessageQue) > 1 {
				messageHandle.MessageQue = messageHandle.MessageQue[1:] // delete the message at index 0 after sending to receiver
			} else {
				messageHandle.MessageQue = []message{}
			}

			messageHandle.mu.Unlock()
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *Server) killSignals() {
	for _, id := range s.unsubscribe {
		logger.InfoLogger.Printf("Killed client: %v", id)

		idd := int32(id)
		m, ok := s.subscribers.Load(idd)
		if !ok && m != nil {
			logger.InfoLogger.Println(fmt.Sprintf("Failed to find subscriber value: %T", idd))
		}
		sub, ok := m.(sub)
		if !ok && m != nil {
			logger.WarningLogger.Panicf("Failed to cast subscriber value: %T", sub)
		}
		if m != nil {
			addToMessageQueue(id, s.lamport.Timestamp, 3, sub.name, "")
		}
		s.subscribers.Delete(id)
		Output(fmt.Sprintf("Client: %v disconnected", id))
	}
}

func (s *Server) Publish(srv protos.ChittyChatService_PublishServer) error {
	logger.InfoLogger.Println("Requests publish")
	er := make(chan error)

	s.lamport.Tick()
	go s.receiveFromStream(srv, er)
	go sendToStream(srv, er)

	return <-er
}

func (s *Server) receiveFromStream(srv protos.ChittyChatService_PublishServer, er_ chan error) {

	//implement a loop
	for {
		mssg, err := srv.Recv()
		if err != nil {
			logger.InfoLogger.Println(fmt.Sprintf("Error occured when recieving message: %v", err))
			break
		}
		id := mssg.ClientId

		switch {
		case mssg.Code == 2: // disconnecting
			s.unsubscribe = append(s.unsubscribe, id)
			s.killSignals()
		case mssg.Code == 1: // chatting
			addToMessageQueue(id, mssg.LamportTimestamp, 2, mssg.UserName, mssg.Msg)
		default:
		}
	}
}

func addToMessageQueue(id, lamport, code int32, username, msg string) {
	messageHandle.mu.Lock()

	messageHandle.MessageQue = append(messageHandle.MessageQue, message{
		ClientUniqueCode: id,
		ClientName:       username,
		Msg:              msg,
		MessageCode:      code,
		Lamport:          lamport,
	})

	logger.InfoLogger.Printf("Message successfully recieved and queued: %v\n", id)

	messageHandle.mu.Unlock()
}

func sendToStream(srv protos.ChittyChatService_PublishServer, er_ chan error) {
	for {
		time.Sleep(500 * time.Millisecond)

		err := srv.Send(&protos.StatusMessage{
			Operation: "Publish()",
			Status:    protos.Status_SUCCESS,
		})

		if err != nil {
			logger.InfoLogger.Println(fmt.Sprintf("An error occured when sending message: %v", err))
			er_ <- err
		}
	}
}

func main() {
	logger.LogFileInit()

	port := 3000

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.InfoLogger.Printf(fmt.Sprintf("FATAL: Connection unable to establish. Failed to listen: %v", err))
	}

	s := &Server{}

	grpcServer := grpc.NewServer()

	protos.RegisterChittyChatServiceServer(grpcServer, s)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.ErrorLogger.Fatalf("FATAL: Server connection failed: %s", err)
		}
	}()


	var o string
	fmt.Scanln(&o)
	os.Exit(3)
}

func Output(input string) {
	fmt.Println(input)
}
