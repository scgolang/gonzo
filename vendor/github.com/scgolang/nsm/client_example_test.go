package nsm

import (
	"context"
	"log"
)

// ExampleClient is a simple example of a non session manager client.
type ExampleClient struct {
	SessionInfo
}

// Open opens a session.
func (c *ExampleClient) Open(info SessionInfo) (string, Error) {
	c.SessionInfo = info
	// Do what you need to do to open the session.
	return "Client has finished opening the session", nil
}

// Save saves a session.
func (c *ExampleClient) Save() (string, Error) {
	// Do what you need to do to save the session.
	return "Client has finished saving the session", nil
}

func Example_client() {
	config := ClientConfig{
		Session: &ExampleClient{},
	}
	c, err := NewClient(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = c.Close() }() // Best effort.

	if err := c.Wait(); err != nil {
		log.Fatal(err)
	}
}
