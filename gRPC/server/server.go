package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/DarkLordOfDeadstiny/Mini-project-2/gRPC"

	"google.golang.org/grpc"
)

type Server struct {
	gRPC.MessageServiceServer
	channel map[string][]chan *gRPC.Message
}

func main() {
	fmt.Println(".:server is starting:.")

	// Create listener tcp on port 5400
	list, err := net.Listen("tcp", "localhost:5400")
	if err != nil {
		log.Fatalf("Failed to listen on port 5400: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	gRPC.RegisterMessageServiceServer(grpcServer, newServer())
	grpcServer.Serve(list)

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to server %v", err)
	}
}

func newServer() *Server {
	s := &Server{
		channel: make(map[string][]chan *gRPC.Message),
	}
	fmt.Println(s)
	return s
}

func (s *Server) JoinChannel(ch *gRPC.Channel, msgStream gRPC.MessageService_JoinChannelServer) error {

	msgChannel := make(chan *gRPC.Message)
	s.channel[ch.ChanName] = append(s.channel[ch.ChanName], msgChannel)

	// doing this never closes the stream
	for {
		select {
		case <-msgStream.Context().Done():
			fmt.Printf("User has disconnected")
			msg := gRPC.Message{
				Channel: &gRPC.Channel{
					ChanName:    ch.ChanName,
					SendersName: ch.SendersName},
				Message:     "some message about disconecting",
				Sender:      ch.SendersName,
				LamportTime: 0,
			}
			msgStream.Send(&msg)
			return nil
		case msg := <-msgChannel:
			fmt.Printf("GO ROUTINE (got message): %v \n", msg)
			msgStream.Send(msg)
		}
	}
}

func (s *Server) SendMessage(msgStream gRPC.MessageService_SendMessageServer) error {
	msg, err := msgStream.Recv()

	if err == io.EOF {
		return nil
	}

	if err != nil {
		return err
	}

	ack := gRPC.MessageAck{Status: "SENT"}
	msgStream.SendAndClose(&ack)

	go func() {
		streams := s.channel[msg.Channel.ChanName]
		for _, msgChan := range streams {
			msgChan <- msg
		}
	}()

	return nil
}
