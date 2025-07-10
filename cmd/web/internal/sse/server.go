package sse

import (
	"io"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	Name  string
	Value any
}

type Client chan Event

func (client Client) Send(name string, value any) {
	client <- Event{name, value}
}

// It keeps a list of clients those are currently attached
// and broadcasting events to those clients.
type Stream struct {
	// Events are pushed to this channel by the main events-gathering routine
	events chan Event

	// New client connections
	NewClients chan Client

	// Closed client connections
	ClosedClients chan Client

	// Total client connections
	TotalClients map[Client]bool

	connectCallback func(client Client)
}

type StreamOption func(stream *Stream)

func OnConnect(callback func(Client)) StreamOption {
	return func(stream *Stream) {
		stream.connectCallback = callback
	}
}

// Initialize event and Start processing requests
func NewStream(options ...StreamOption) *Stream {
	stream := &Stream{
		events:        make(chan Event),
		NewClients:    make(chan Client),
		ClosedClients: make(chan Client),
		TotalClients:  make(map[Client]bool),
	}

	for _, opt := range options {
		opt(stream)
	}

	go stream.listen()

	return stream
}

// It Listens all incoming requests from clients.
// Handles addition and removal of clients and broadcast messages to clients.
func (stream *Stream) listen() {
	for {
		select {
		// Add new available client
		case client := <-stream.NewClients:
			stream.TotalClients[client] = true
			log.Printf("Client added. %d registered clients", len(stream.TotalClients))

		// Remove closed client
		case client := <-stream.ClosedClients:
			delete(stream.TotalClients, client)
			close(client)
			log.Printf("Removed client. %d registered clients", len(stream.TotalClients))

		// Broadcast message to client
		case eventMsg := <-stream.events:
			for clientMessageChan := range stream.TotalClients {
				select {
				case clientMessageChan <- eventMsg:
					// Message sent successfully
				default:
					// Failed to send, dropping message
				}
			}
		}
	}
}

// New event messages are broadcast to all registered client connection channels
func (stream *Stream) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set headers
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")

		// Initialize client channel
		client := make(Client)
		if stream.connectCallback != nil {
			go stream.connectCallback(client)
		}

		// Send new connection to event stream
		stream.NewClients <- client

		go func() {
			<-c.Writer.CloseNotify()

			// Send closed connection to event stream
			stream.ClosedClients <- client

			// Drain client channel so that it does not block. Server may keep sending messages to this channel
			for range client {
			}

		}()

		c.Stream(func(w io.Writer) bool {
			// Stream message to client from message channel
			if msg, ok := <-client; ok {
				c.SSEvent(msg.Name, msg.Value)
				return true
			}
			return false
		})
	}
}

func (stream *Stream) Send(name string, value any) {
	stream.events <- Event{name, value}
}
