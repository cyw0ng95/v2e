package rpc

import (
	"context"

	"github.com/cyw0ng95/v2e/pkg/notes"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// Client provides RPC-based access to notes services
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

	return &Client{
		subprocess:  sp,
	}
}

// InitializeClient initializes the RPC client with service implementations
func InitializeClient(target string, bookmarkService notes.BookmarkServiceInterface, noteService notes.NoteServiceInterface, memoryCardService notes.MemoryCardServiceInterface, crossReferenceService notes.CrossReferenceServiceInterface, historyService notes.HistoryServiceInterface) *Client {
	// Create a subprocess client that can communicate with the target service
	sp := subprocess.New(target)

	client := &Client{
		subprocess: sp,
	}

	client.BookmarkService = bookmarkService
	client.NoteService = noteService
	client.MemoryCardService = memoryCardService
	client.CrossReferenceService = crossReferenceService
	client.HistoryService = historyService

	return client
}

// Connect establishes connections to remote services
func (c *Client) Connect(ctx context.Context) error {
	// In a real implementation, this would establish connections to remote services
	// For now, we'll just return nil
	return nil
}

// Close closes connections to remote services
func (c *Client) Close() error {
	// In a real implementation, this would close connections
	// For now, we'll just return nil
	return nil
}
