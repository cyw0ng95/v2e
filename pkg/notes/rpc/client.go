package rpc

import (
	"context"

	"github.com/cyw0ng95/v2e/pkg/notes"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// Client provides RPC-based access to the notes services
type Client struct {
	subprocess            *subprocess.Subprocess
	BookmarkService       notes.BookmarkServiceInterface
	NoteService           notes.NoteServiceInterface
	MemoryCardService     notes.MemoryCardServiceInterface
	CrossReferenceService notes.CrossReferenceServiceInterface
	HistoryService        notes.HistoryServiceInterface
}

// NewClient creates a new RPC client for the notes services
func NewClient(target string) *Client {
	// Create a subprocess client that can communicate with the target service
	sp := subprocess.New(target)

	client := &Client{
		subprocess: sp,
	}

	// Initialize RPC service implementations
	rpcClient := notes.NewRPCClient(sp)
	serviceContainer := notes.NewRPCServiceContainer(rpcClient)

	client.BookmarkService = serviceContainer.BookmarkService
	client.NoteService = serviceContainer.NoteService
	client.MemoryCardService = serviceContainer.MemoryCardService
	client.CrossReferenceService = serviceContainer.CrossReferenceService
	client.HistoryService = serviceContainer.HistoryService

	return client
}

// Connect establishes connections to the remote services
func (c *Client) Connect(ctx context.Context) error {
	// In a real implementation, this would establish connections to remote services
	// For now, we'll just return nil
	return nil
}

// Close closes the connections to the remote services
func (c *Client) Close() error {
	// In a real implementation, this would close connections
	// For now, we'll just return nil
	return nil
}
