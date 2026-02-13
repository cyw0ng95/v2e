package storage

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/graph"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// jsonBufferPool is a sync.Pool for reusing JSON buffers
var jsonBufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 1024) // 1KB buffers for JSON
		return &buf
	},
}

// Bucket names for graph data
var (
	BucketNodes         = []byte("graph_nodes")           // Node data
	BucketEdges         = []byte("graph_edges")           // Edge data
	BucketMetadata      = []byte("graph_metadata")        // Graph metadata (stats, timestamps)
	BucketCheckpoint    = []byte("graph_checkpoint")       // Checkpoint data
	BucketIncremental   = []byte("graph_incremental")     // Incremental changes
	BucketCheckpointLog = []byte("graph_checkpoint_log")   // Checkpoint log for recovery
)

// GraphMetadata stores metadata about persisted graph
type GraphMetadata struct {
	NodeCount      int       `json:"node_count"`
	EdgeCount      int       `json:"edge_count"`
	LastSaved      time.Time `json:"last_saved"`
	LastCheckpoint string    `json:"last_checkpoint,omitempty"`
	Version        string    `json:"version"`
	CheckpointID   int64     `json:"checkpoint_id"`
	BaseCheckpoint string    `json:"base_checkpoint,omitempty"`
}

// IncrementalChange represents a single incremental change
type IncrementalChange struct {
	Type     string                 `json:"type"`     // "add_node", "remove_node", "add_edge", "remove_edge"
	NodeURN  string                 `json:"node_urn,omitempty"`
	From     string                 `json:"from,omitempty"`
	To       string                 `json:"to,omitempty"`
	EdgeType string                 `json:"edge_type,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// CheckpointData represents a checkpoint with incremental changes
type CheckpointData struct {
	ID          int64               `json:"id"`
	ParentID    int64               `json:"parent_id"`    // Parent checkpoint for incremental recovery
	Timestamp   time.Time           `json:"timestamp"`
	NodeCount   int                 `json:"node_count"`
	EdgeCount   int                 `json:"edge_count"`
	Changes     []IncrementalChange `json:"changes,omitempty"`
	IsFull      bool                `json:"is_full"`      // True if this is a full checkpoint
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
	db           *bolt.DB
	logger        *common.Logger
	nextCheckID   int64
	checkpointMu  sync.Mutex
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
			BucketIncremental,
			BucketCheckpointLog,
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

	// Get buffer from pool
		buf := jsonBufferPool.Get().(*[]byte)
		defer jsonBufferPool.Put(buf)

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

			// Use sequential ID as key for edges (optimized allocation)
			key := []byte("edge_" + strconv.FormatInt(int64(i), 10))
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

// SaveIncrementalCheckpoint saves an incremental checkpoint with only changes
// Returns the checkpoint ID for recovery
func (s *GraphStore) SaveIncrementalCheckpoint(g *graph.Graph, changes []IncrementalChange, parentID int64) (int64, error) {
	s.checkpointMu.Lock()
	s.nextCheckID++
	checkID := s.nextCheckID
	s.checkpointMu.Unlock()

	startTime := time.Now()

	err := s.db.Update(func(tx *bolt.Tx) error {
		// Save checkpoint data
		checkpointData := CheckpointData{
			ID:        checkID,
			ParentID:  parentID,
			Timestamp: time.Now(),
			NodeCount: g.NodeCount(),
			EdgeCount: g.EdgeCount(),
			Changes:   changes,
			IsFull:    false,
		}

		data, err := sonic.Marshal(checkpointData)
		if err != nil {
			return fmt.Errorf("failed to marshal checkpoint: %w", err)
		}

		// Use checkpoint ID as key
		key := []byte(fmt.Sprintf("chk_%d", checkID))
		b := tx.Bucket(BucketCheckpointLog)
		if err := b.Put(key, data); err != nil {
			return fmt.Errorf("failed to save checkpoint: %w", err)
		}

		// Update metadata with latest checkpoint
		return s.updateCheckpointMetadata(tx, checkID, fmt.Sprintf("chk_%d", checkID))
	})

	if err != nil {
		return 0, err
	}

	duration := time.Since(startTime)
	if s.logger != nil {
		s.logger.Info("Incremental checkpoint %d saved in %v (%d changes)", checkID, duration, len(changes))
	}

	return checkID, nil
}

// SaveFullCheckpoint saves a full checkpoint (snapshot of entire graph)
// This should be called periodically to establish new base checkpoints
func (s *GraphStore) SaveFullCheckpoint(g *graph.Graph) (int64, error) {
	s.checkpointMu.Lock()
	s.nextCheckID++
	checkID := s.nextCheckID
	s.checkpointMu.Unlock()

	startTime := time.Now()

	err := s.db.Update(func(tx *bolt.Tx) error {
		// Save full graph data to main buckets
		if err := s.saveGraphToTx(tx, g); err != nil {
			return err
		}

		// Create checkpoint metadata
		checkpointData := CheckpointData{
			ID:        checkID,
			ParentID:  0, // Full checkpoint has no parent
			Timestamp: time.Now(),
			NodeCount: g.NodeCount(),
			EdgeCount: g.EdgeCount(),
			IsFull:    true,
		}

		data, err := sonic.Marshal(checkpointData)
		if err != nil {
			return fmt.Errorf("failed to marshal checkpoint: %w", err)
		}

		key := []byte(fmt.Sprintf("chk_%d", checkID))
		b := tx.Bucket(BucketCheckpointLog)
		if err := b.Put(key, data); err != nil {
			return fmt.Errorf("failed to save checkpoint: %w", err)
		}

		// Update metadata
		return s.updateCheckpointMetadata(tx, checkID, fmt.Sprintf("chk_%d", checkID))
	})

	if err != nil {
		return 0, err
	}

	duration := time.Since(startTime)
	if s.logger != nil {
		s.logger.Info("Full checkpoint %d saved in %v (nodes: %d, edges: %d)",
			checkID, duration, g.NodeCount(), g.EdgeCount())
	}

	return checkID, nil
}

// LoadFromCheckpoint loads graph from a specific checkpoint
// For incremental checkpoints, it will load the parent checkpoint and apply changes
func (s *GraphStore) LoadFromCheckpoint(checkID int64) (*graph.Graph, error) {
	startTime := time.Now()

	var resultGraph *graph.Graph
	var checkpointData CheckpointData

	// First, load the checkpoint data
	err := s.db.View(func(tx *bolt.Tx) error {
		key := []byte(fmt.Sprintf("chk_%d", checkID))
		b := tx.Bucket(BucketCheckpointLog)
		data := b.Get(key)
		if data == nil {
			return fmt.Errorf("checkpoint %d not found", checkID)
		}

		if err := sonic.Unmarshal(data, &checkpointData); err != nil {
			return fmt.Errorf("failed to unmarshal checkpoint: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// If this is a full checkpoint or has no parent, load from main buckets
	if checkpointData.IsFull || checkpointData.ParentID == 0 {
		g, err := s.LoadGraph()
		if err != nil {
			return nil, err
		}
		resultGraph = g
	} else {
		// Recursive: load parent checkpoint first
		parentGraph, err := s.LoadFromCheckpoint(checkpointData.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to load parent checkpoint %d: %w", checkpointData.ParentID, err)
		}
		resultGraph = parentGraph

		// Apply incremental changes
		if err := s.applyChanges(resultGraph, checkpointData.Changes); err != nil {
			return nil, fmt.Errorf("failed to apply changes: %w", err)
		}
	}

	duration := time.Since(startTime)
	if s.logger != nil {
		s.logger.Info("Graph loaded from checkpoint %d in %v (nodes: %d, edges: %d)",
			checkID, duration, resultGraph.NodeCount(), resultGraph.EdgeCount())
	}

	return resultGraph, nil
}

// GetLatestCheckpointID returns the ID of the most recent checkpoint
func (s *GraphStore) GetLatestCheckpointID() (int64, error) {
	var latestID int64 = 0

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketCheckpointLog)
		if b == nil {
			return nil
		}

		c := b.Cursor()
		k, _ := c.Last()
		if k != nil {
			// Parse key "chk_<id>"
			var id int64
			_, err := fmt.Sscanf(string(k), "chk_%d", &id)
			if err == nil {
				latestID = id
			}
		}

		return nil
	})

	return latestID, err
}

// ListCheckpoints returns a list of all checkpoint IDs with timestamps
func (s *GraphStore) ListCheckpoints() ([]CheckpointData, error) {
	var checkpoints []CheckpointData

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketCheckpointLog)
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			var cp CheckpointData
			if err := sonic.Unmarshal(v, &cp); err != nil {
				return err
			}
			checkpoints = append(checkpoints, cp)
			return nil
		})
	})

	return checkpoints, err
}

// DeleteCheckpoint removes a checkpoint and its metadata
func (s *GraphStore) DeleteCheckpoint(checkID int64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		key := []byte(fmt.Sprintf("chk_%d", checkID))
		b := tx.Bucket(BucketCheckpointLog)
		return b.Delete(key)
	})
}

// CompactCheckpoints removes old checkpoints, keeping only the most recent N
func (s *GraphStore) CompactCheckpoints(keep int) error {
	checkpoints, err := s.ListCheckpoints()
	if err != nil {
		return err
	}

	if len(checkpoints) <= keep {
		return nil // Nothing to do
	}

	// Sort by ID (which correlates with time)
	// Remove oldest checkpoints
	toDelete := len(checkpoints) - keep
	for i := 0; i < toDelete; i++ {
		if err := s.DeleteCheckpoint(checkpoints[i].ID); err != nil {
			return err
		}
	}

	if s.logger != nil {
		s.logger.Info("Compacted checkpoints: deleted %d, kept %d", toDelete, keep)
	}

	return nil
}

// RecoverFromLatestCheckpoint attempts to recover from the most recent checkpoint
func (s *GraphStore) RecoverFromLatestCheckpoint() (*graph.Graph, int64, error) {
	checkID, err := s.GetLatestCheckpointID()
	if err != nil {
		return nil, 0, err
	}

	if checkID == 0 {
		return nil, 0, fmt.Errorf("no checkpoints found")
	}

	g, err := s.LoadFromCheckpoint(checkID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to load checkpoint %d: %w", checkID, err)
	}

	return g, checkID, nil
}

// saveGraphToTx saves graph to transaction (internal helper)
func (s *GraphStore) saveGraphToTx(tx *bolt.Tx, g *graph.Graph) error {
	// Clear existing data
	if err := s.clearBucket(tx, BucketNodes); err != nil {
		return err
	}
	if err := s.clearBucket(tx, BucketEdges); err != nil {
		return err
	}

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

		key := []byte("edge_" + strconv.FormatInt(int64(i), 10))
		if err := edgesBucket.Put(key, data); err != nil {
			return fmt.Errorf("failed to save edge: %w", err)
		}
		edgeCount++
	}

	// Update metadata
	metadata := GraphMetadata{
		NodeCount: nodeCount,
		EdgeCount: edgeCount,
		LastSaved: time.Now(),
		Version:   "1.0",
	}
	return s.saveMetadata(tx, metadata)
}

// updateCheckpointMetadata updates the checkpoint reference in metadata
func (s *GraphStore) updateCheckpointMetadata(tx *bolt.Tx, checkID int64, checkKey string) error {
	metadata := GraphMetadata{
		CheckpointID:   checkID,
		LastCheckpoint: checkKey,
		LastSaved:      time.Now(),
	}
	b := tx.Bucket(BucketMetadata)
	data, err := sonic.Marshal(metadata)
	if err != nil {
		return err
	}
	return b.Put([]byte("checkpoint_meta"), data)
}

// applyChanges applies incremental changes to a graph
func (s *GraphStore) applyChanges(g *graph.Graph, changes []IncrementalChange) error {
	for _, change := range changes {
		switch change.Type {
		case "add_node":
			if change.NodeURN == "" {
				continue
			}
			u, err := urn.Parse(change.NodeURN)
			if err != nil {
				return fmt.Errorf("failed to parse URN %s: %w", change.NodeURN, err)
			}
			// Add node (no-op if exists)
			g.AddNode(u, change.Data)

		case "remove_node":
			// Note: graph.Graph doesn't support remove operation yet
			// Skip with a warning
			if s.logger != nil {
				s.logger.Warn("remove_node operation not supported, skipping")
			}

		case "add_edge":
			if change.From == "" || change.To == "" {
				continue
			}
			from, err := urn.Parse(change.From)
			if err != nil {
				return fmt.Errorf("failed to parse from URN %s: %w", change.From, err)
			}
			to, err := urn.Parse(change.To)
			if err != nil {
				return fmt.Errorf("failed to parse to URN %s: %w", change.To, err)
			}
			// Add edge (returns error if it exists, which we ignore)
			_ = g.AddEdge(from, to, graph.EdgeType(change.EdgeType), change.Data)

		case "remove_edge":
			// Note: graph.Graph doesn't support remove operation yet
			// Skip with a warning
			if s.logger != nil {
				s.logger.Warn("remove_edge operation not supported, skipping")
			}
		}
	}
	return nil
}
