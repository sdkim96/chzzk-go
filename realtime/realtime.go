// package realtime provides an interface for real-time communication with a server.
// It defines the RealTime interface, which includes methods for dialing a connection,
// running a loop to handle incoming and outgoing messages, and closing the connection.
package realtime

import "context"

type Realtime interface {
	Dial(ctx context.Context, url string) error
	Loop(ctx context.Context, recv chan []byte, send chan []byte, errCh chan error)
	Close(ctx context.Context) error
}
