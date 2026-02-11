package storage

import (
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/graph"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// Bucket names for graph data
var (
	BucketNodes      = []byte("graph_nodes")      // Node data
	BucketEdges      = []byte("graph_edges")      // Edge data
	BucketMetadata   = []byte("graph_metadata")   // Graph metadata (stats, timestamps)
	BucketCheckpoint = []byte("graph_checkpoint") // Checkpoint data
)

// GraphMetadata stores metadata about the persisted graph
type GraphMetadata struct {
	NodeCount      int       `json:"node_count"`
	EdgeCount      int       `json:"edge_count"`
	LastSaved      time.Time `json:"last_saved"`
	LastCheckpoint string    `json:"last_checkpoint,omitempty"`
	Version        string    `json:"version"`
}

// NodeData represents a serialized node
type NodeData struct {
	URN        string                 `json:"urn"`
	Properties map[string]interface{} `json:"properties"`
}

// EdgeData represents a serialized edge
type EdgeData struct {
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

// GraphStore provides BoltDB-based graph persistence
type GraphStore struct {
	db     *bolt.DB
	logger *common.Logger
}

// NewGraphStore creates a new graph storage instance
func NewGraphStore(dbPath string, logger *common.Logger) (*GraphStore, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open graph database: %w", err)
	}

	// Create all buckets
	err = db.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			BucketNodes,
			BucketEdges,
			BucketMetadata,
			BucketCheckpoint,
		}
		for _, bucketName := range buckets {
			if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &GraphStore{
		db:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (s *GraphStore) Close() error {
	return s.db.Close()
}

// SaveGraph saves the entire graph to BoltDB
func (s *GraphStore) SaveGraph(g *graph.Graph) error {
	startTime := time.Now()

	err := s.db.Update(func(tx *bolt.Tx) error {
		// Clear existing data
		if err := s.clearBucket(tx, BucketNodes); err != nil {
			return err
		}
		if err := s.clearBucket(tx, BucketEdges); err != nil {
			return err
		}

		// Get buckets after clearing
		nodesBucket := tx.Bucket(BucketNodes)
		edgesBucket := tx.Bucket(BucketEdges)

		// Save all nodes
		nodeCount := 0
		nodes := g.GetAllNodes()
		for _, node := range nodes {
			nodeData := NodeData{
				URN:        node.URN.String(),
				Properties: node.Properties,
			}

			data, err := sonic.Marshal(nodeData)
			if err != nil {
				return fmt.Errorf("failed to marshal node %s: %w", node.URN.String(), err)
			}

			key := []byte(node.URN.Key())
			if err := nodesBucket.Put(key, data); err != nil {
				return fmt.Errorf("failed to save node %s: %w", node.URN.String(), err)
			}
			nodeCount++
		}

		// Save all edges
		edgeCount := 0
		edges := g.GetAllEdges()
		for i, edge := range edges {
			edgeData := EdgeData{
				From:       edge.From.String(),
				To:         edge.To.String(),
				Type:       string(edge.Type),
				Properties: edge.Properties,
			}

			data, err := sonic.Marshal(edgeData)
			if err != nil {
				return fmt.Errorf("failed to marshal edge: %w", err)
			}

			// Use sequential ID as key for edges
			key := []byte(fmt.Sprintf("edge_%d", i))
			if err := edgesBucket.Put(key, data); err != nil {
				return fmt.Errorf("failed to save edge: %w", err)
			}
			edgeCount++
		}

		// Save metadata
		metadata := GraphMetadata{
			NodeCount: nodeCount,
			EdgeCount: edgeCount,
			LastSaved: time.Now(),
			Version:   "1.0",
		}
		return s.saveMetadata(tx, metadata)
	})

	if err != nil {
		return err
	}

	duration := time.Since(startTime)
	if s.logger != nil {
		s.logger.Info("Graph saved to disk in %v", duration)
	}

	return nil
}

// LoadGraph loads the entire graph from BoltDB
func (s *GraphStore) LoadGraph() (*graph.Graph, error) {
	startTime := time.Now()
	g := graph.New()

	err := s.db.View(func(tx *bolt.Tx) error {
		nodesBucket := tx.Bucket(BucketNodes)
		edgesBucket := tx.Bucket(BucketEdges)

		// Load all nodes
		err := nodesBucket.ForEach(func(k, v []byte) error {
			var nodeData NodeData
			if err := sonic.Unmarshal(v, &nodeData); err != nil {
				return fmt.Errorf("failed to unmarshal node: %w", err)
			}

			u, err := urn.Parse(nodeData.URN)
			if err != nil {
				return fmt.Errorf("failed to parse URN %s: %w", nodeData.URN, err)
			}

			g.AddNode(u, nodeData.Properties)
			return nil
		})
		if err != nil {
			return err
		}

		// Load all edges
		err = edgesBucket.ForEach(func(k, v []byte) error {
			var edgeData EdgeData
			if err := sonic.Unmarshal(v, &edgeData); err != nil {
				return fmt.Errorf("failed to unmarshal edge: %w", err)
			}

			from, err := urn.Parse(edgeData.From)
			if err != nil {
				return fmt.Errorf("failed to parse from URN %s: %w", edgeData.From, err)
			}

			to, err := urn.Parse(edgeData.To)
			if err != nil {
				return fmt.Errorf("failed to parse to URN %s: %w", edgeData.To, err)
			}

			edgeType := graph.EdgeType(edgeData.Type)
			if err := g.AddEdge(from, to, edgeType, edgeData.Properties); err != nil {
				// Log error but continue loading
				if s.logger != nil {
					s.logger.Warn("Failed to add edge: %v", err)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)
	if s.logger != nil {
		s.logger.Info("Graph loaded from disk in %v (nodes: %d, edges: %d)",
			duration, g.NodeCount(), g.EdgeCount())
	}

	return g, nil
}

// GetMetadata retrieves the graph metadata
func (s *GraphStore) GetMetadata() (*GraphMetadata, error) {
	var metadata GraphMetadata

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketMetadata)
		v := b.Get([]byte("metadata"))
		if v == nil {
			return fmt.Errorf("metadata not found")
		}
		return sonic.Unmarshal(v, &metadata)
	})

	if err != nil {
		return nil, err
	}

	return &metadata, nil
}

// SaveCheckpoint saves a checkpoint marker
func (s *GraphStore) SaveCheckpoint(checkpointID string, data map[string]interface{}) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketCheckpoint)

		checkpointData := map[string]interface{}{
			"id":        checkpointID,
			"timestamp": time.Now(),
			"data":      data,
		}

		jsonData, err := sonic.Marshal(checkpointData)
		if err != nil {
			return err
		}

		return b.Put([]byte(checkpointID), jsonData)
	})
}

// GetCheckpoint retrieves a checkpoint
func (s *GraphStore) GetCheckpoint(checkpointID string) (map[string]interface{}, error) {
	var result map[string]interface{}

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketCheckpoint)
		v := b.Get([]byte(checkpointID))
		if v == nil {
			return fmt.Errorf("checkpoint not found: %s", checkpointID)
		}
		return sonic.Unmarshal(v, &result)
	})

	return result, err
}

// ClearGraph removes all graph data from storage
func (s *GraphStore) ClearGraph() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if err := s.clearBucket(tx, BucketNodes); err != nil {
			return err
		}
		if err := s.clearBucket(tx, BucketEdges); err != nil {
			return err
		}
		if err := s.clearBucket(tx, BucketMetadata); err != nil {
			return err
		}
		return nil
	})
}

// clearBucket removes all keys from a bucket
func (s *GraphStore) clearBucket(tx *bolt.Tx, bucketName []byte) error {
	// Delete the bucket
	if err := tx.DeleteBucket(bucketName); err != nil && err != bolt.ErrBucketNotFound {
		return err
	}
	// Recreate it empty
	_, err := tx.CreateBucket(bucketName)
	return err
}

// saveMetadata saves metadata to the database
func (s *GraphStore) saveMetadata(tx *bolt.Tx, metadata GraphMetadata) error {
	b := tx.Bucket(BucketMetadata)

	data, err := sonic.Marshal(metadata)
	if err != nil {
		return err
	}

	return b.Put([]byte("metadata"), data)
}

// Stats returns storage statistics
func (s *GraphStore) Stats() (*bolt.Stats, error) {
	stats := s.db.Stats()
	return &stats, nil
}
