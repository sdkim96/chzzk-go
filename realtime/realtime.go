// package realtime provides an interface for real-time communication with a server.
// It defines the RealTime interface, which includes methods for dialing a connection,
// running a loop to handle incoming and outgoing messages, and closing the connection.
package realtime

import "context"

type RealTime interface {
	Dial(ctx context.Context, url string) error
	Loop(ctx context.Context, recv <-chan []byte, send chan<- []byte, err chan<- error)
	Close(ctx context.Context) error
}

// Run is a helper function that sets up the real-time connection and starts the loop in a separate goroutine.
func Run(ctx context.Context, rt RealTime, url string) (<-chan []byte, chan<- []byte, <-chan error, error) {
	if err := rt.Dial(ctx, url); err != nil {
		return nil, nil, nil, err
	}
	recv, send, errChan := make(chan []byte), make(chan []byte), make(chan error, 1)
	go func() {
		defer rt.Close(ctx)
		rt.Loop(ctx, recv, send, errChan)
	}()
	return recv, send, errChan, nil
}
