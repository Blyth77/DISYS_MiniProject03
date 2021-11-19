package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	logger "github.com/Blyth77/DISYS_MiniProject03/logger"
	protos "github.com/Blyth77/DISYS_MiniProject03/proto"

	"google.golang.org/grpc"
)

var (
	serverId    int32
	port = 3000
)


type message struct {
	ClientUniqueCode int32
	ClientName       string
	Msg              string
	MessageCode      int32
}

type raw struct {
	MessageQue []message
	mu         sync.Mutex
}

type Server struct {
	protos.UnimplementedAuctionhouseServiceServer
	buyers      sync.Map
	unsubscribe []int32
}

type sub struct {
	stream   protos.AuctionhouseService_BroadcastServer
	finished chan<- bool
	name     string
}

var messageHandle = raw{}

func (s *Server) Broadcast(request *protos.Subscription, stream protos.AuctionhouseService_BroadcastServer) error {

	fin := make(chan bool)

	s.buyers.Store(request.ClientId, sub{stream: stream, finished: fin, name: request.UserName})

	addToMessageQueue(request.ClientId, 1, request.UserName, "")

	go s.sendToClients(stream)

	bl := make(chan error)
	return <-bl
}

func (s *Server) sendToClients(srv protos.AuctionhouseService_BroadcastServer) {
	for {
		for {

			time.Sleep(500 * time.Millisecond)

			messageHandle.mu.Lock()

			if len(messageHandle.MessageQue) == 0 {
				messageHandle.mu.Unlock()
				break
			}
			senderUniqueCode := messageHandle.MessageQue[0].ClientUniqueCode
			senderName := messageHandle.MessageQue[0].ClientName
			messageFromServer := messageHandle.MessageQue[0].Msg
			messageCode := messageHandle.MessageQue[0].MessageCode

			messageHandle.mu.Unlock()

			s.buyers.Range(func(k, v interface{}) bool {
				id, ok := k.(int32)
				if !ok {
					logger.InfoLogger.Println(fmt.Sprintf("Failed to cast buyers key: %T", k))
					return false
				}
				sub, ok := v.(sub)
				if !ok {
					logger.InfoLogger.Println(fmt.Sprintf("Failed to cast buyers value: %T", v))
					return false
				}
				// Send data over the gRPC stream to the client
				if err := sub.stream.Send(&protos.ChatRoomMessages{
					Msg:      messageFromServer,
					Username: senderName,
					ClientId: senderUniqueCode,
					Code:     messageCode,
				}); err != nil {
					logger.ErrorLogger.Output(2, (fmt.Sprintf("Failed to send data to client: %v", err)))
					s.unsubscribe = append(s.unsubscribe, id)
				}
				return true
			})

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
		//logger.InfoLogger.Printf("Killed client: %v", id)

		idd := int32(id)
		m, ok := s.buyers.Load(idd)
		if !ok && m != nil {
			logger.InfoLogger.Println(fmt.Sprintf("Failed to find buyer value: %T", idd))
		}
		sub, ok := m.(sub)
		if !ok && m != nil {
			logger.WarningLogger.Panicf("Failed to cast buyer value: %T", sub)
		}
		if m != nil {
			addToMessageQueue(id, 3, sub.name, "")
		}
		s.buyers.Delete(id)
	}
}

func (s *Server) Publish(srv protos.AuctionhouseService_PublishServer) error {
	er := make(chan error)

	go s.receiveFromStream(srv, er)
	go sendToStream(srv, er)

	return <-er
}

func (s *Server) receiveFromStream(srv protos.AuctionhouseService_PublishServer, er_ chan error) {

	//implement a loop
	for {
		mssg, err := srv.Recv()
		if err != nil {
			break
		}
		id := mssg.ClientId

		switch {
		case mssg.Code == 2: // disconnecting
			s.unsubscribe = append(s.unsubscribe, id)
			s.killSignals()
		case mssg.Code == 1: // chatting
			addToMessageQueue(id, 2, mssg.UserName, mssg.Msg)
		default:
		}
	}
}

func addToMessageQueue(id, code int32, username, msg string) {
	messageHandle.mu.Lock()

	messageHandle.MessageQue = append(messageHandle.MessageQue, message{
		ClientUniqueCode: id,
		ClientName:       username,
		Msg:              msg,
		MessageCode:      code,
	})

	logger.InfoLogger.Printf("Message successfully recieved and queued: %v\n", id)

	messageHandle.mu.Unlock()
}

func sendToStream(srv protos.AuctionhouseService_PublishServer, er_ chan error) {
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
	serverId = 1
	logger.LogFileInit("server", serverId)

	s := &Server{}


	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.InfoLogger.Printf(fmt.Sprintf("FATAL: Connection unable to establish. Failed to listen: %v", err))
	}


	grpcServer := grpc.NewServer()

	protos.RegisterAuctionhouseServiceServer(grpcServer, s)

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
