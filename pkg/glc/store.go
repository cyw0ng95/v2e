package glc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Store provides GLC data storage operations
type Store struct {
	db *gorm.DB
}

// NewStore creates a new GLC store
func NewStore(db *gorm.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}
	return &Store{db: db}, nil
}

// Close closes the database connection (no-op for shared connection)
func (s *Store) Close() error {
	return nil
}

// ============================================================================
// Graph Operations
// ============================================================================

// CreateGraph creates a new graph
func (s *Store) CreateGraph(ctx context.Context, name, description, presetID string, nodes, edges, viewport string) (*GraphModel, error) {
	graph := &GraphModel{
		GraphID:     uuid.New().String(),
		Name:        name,
		Description: description,
		PresetID:    presetID,
		Nodes:       nodes,
		Edges:       edges,
		Viewport:    viewport,
		Version:     1,
	}

	if err := s.db.WithContext(ctx).Create(graph).Error; err != nil {
		return nil, fmt.Errorf("failed to create graph: %w", err)
	}

	return graph, nil
}

// GetGraph retrieves a graph by graph_id
func (s *Store) GetGraph(ctx context.Context, graphID string) (*GraphModel, error) {
	var graph GraphModel
	if err := s.db.WithContext(ctx).Where("graph_id = ?", graphID).First(&graph).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("graph not found: %s", graphID)
		}
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}
	return &graph, nil
}

// GetGraphByDBID retrieves a graph by database ID
func (s *Store) GetGraphByDBID(ctx context.Context, id uint) (*GraphModel, error) {
	var graph GraphModel
	if err := s.db.WithContext(ctx).First(&graph, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("graph not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}
	return &graph, nil
}

// UpdateGraph updates a graph and creates a version snapshot
func (s *Store) UpdateGraph(ctx context.Context, graphID string, updates map[string]interface{}) (*GraphModel, error) {
	var graph GraphModel
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("graph_id = ?", graphID).First(&graph).Error; err != nil {
			return err
		}

		// Create version snapshot before update (if nodes/edges changed)
		if _, ok := updates["nodes"]; ok || updates["edges"] != nil {
			version := &GraphVersionModel{
				GraphID: graph.ID,
				Version: graph.Version,
				Nodes:   graph.Nodes,
				Edges:   graph.Edges,
				Viewport: graph.Viewport,
			}
			if err := tx.Create(version).Error; err != nil {
				return fmt.Errorf("failed to create version snapshot: %w", err)
			}
		}

		// Apply updates
		if err := tx.Model(&graph).Updates(updates).Error; err != nil {
			return err
		}

		// Increment version
		return tx.Model(&graph).Update("version", gorm.Expr("version + 1")).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update graph: %w", err)
	}

	// Reload the graph
	if err := s.db.WithContext(ctx).Where("graph_id = ?", graphID).First(&graph).Error; err != nil {
		return nil, err
	}

	return &graph, nil
}

// DeleteGraph soft-deletes a graph
func (s *Store) DeleteGraph(ctx context.Context, graphID string) error {
	result := s.db.WithContext(ctx).Where("graph_id = ?", graphID).Delete(&GraphModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete graph: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("graph not found: %s", graphID)
	}
	return nil
}

// ListGraphs lists graphs with pagination
func (s *Store) ListGraphs(ctx context.Context, presetID string, offset, limit int) ([]GraphModel, int64, error) {
	var graphs []GraphModel
	var total int64

	query := s.db.WithContext(ctx).Model(&GraphModel{})
	if presetID != "" {
		query = query.Where("preset_id = ?", presetID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count graphs: %w", err)
	}

	if err := query.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&graphs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list graphs: %w", err)
	}

	return graphs, total, nil
}

// ListRecentGraphs lists recently accessed graphs
func (s *Store) ListRecentGraphs(ctx context.Context, limit int) ([]GraphModel, error) {
	var graphs []GraphModel
	if err := s.db.WithContext(ctx).Order("updated_at DESC").Limit(limit).Find(&graphs).Error; err != nil {
		return nil, fmt.Errorf("failed to list recent graphs: %w", err)
	}
	return graphs, nil
}

// ============================================================================
// Version Operations
// ============================================================================

// GetVersion retrieves a specific version of a graph
func (s *Store) GetVersion(ctx context.Context, graphDBID uint, version int) (*GraphVersionModel, error) {
	var v GraphVersionModel
	if err := s.db.WithContext(ctx).Where("graph_id = ? AND version = ?", graphDBID, version).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("version not found: graph=%d, version=%d", graphDBID, version)
		}
		return nil, fmt.Errorf("failed to get version: %w", err)
	}
	return &v, nil
}

// ListVersions lists all versions of a graph
func (s *Store) ListVersions(ctx context.Context, graphDBID uint, limit int) ([]GraphVersionModel, error) {
	var versions []GraphVersionModel
	query := s.db.WithContext(ctx).Where("graph_id = ?", graphDBID).Order("version DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}
	return versions, nil
}

// RestoreVersion restores a graph to a specific version
func (s *Store) RestoreVersion(ctx context.Context, graphID string, version int) (*GraphModel, error) {
	var graph GraphModel
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("graph_id = ?", graphID).First(&graph).Error; err != nil {
			return err
		}

		var v GraphVersionModel
		if err := tx.Where("graph_id = ? AND version = ?", graph.ID, version).First(&v).Error; err != nil {
			return err
		}

		// Restore nodes, edges, viewport
		return tx.Model(&graph).Updates(map[string]interface{}{
			"nodes":    v.Nodes,
			"edges":    v.Edges,
			"viewport": v.Viewport,
		}).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to restore version: %w", err)
	}

	// Reload
	if err := s.db.WithContext(ctx).Where("graph_id = ?", graphID).First(&graph).Error; err != nil {
		return nil, err
	}

	return &graph, nil
}

// DeleteOldVersions deletes versions older than a certain count (keeping recent N)
func (s *Store) DeleteOldVersions(ctx context.Context, graphDBID uint, keepCount int) error {
	// Get version IDs to keep
	var keepIDs []uint
	if err := s.db.WithContext(ctx).
		Model(&GraphVersionModel{}).
		Where("graph_id = ?", graphDBID).
		Order("version DESC").
		Limit(keepCount).
		Pluck("id", &keepIDs).Error; err != nil {
		return fmt.Errorf("failed to get versions to keep: %w", err)
	}

	if len(keepIDs) == 0 {
		return nil
	}

	// Delete older versions
	if err := s.db.WithContext(ctx).
		Where("graph_id = ? AND id NOT IN ?", graphDBID, keepIDs).
		Delete(&GraphVersionModel{}).Error; err != nil {
		return fmt.Errorf("failed to delete old versions: %w", err)
	}

	return nil
}

// ============================================================================
// User Preset Operations
// ============================================================================

// CreateUserPreset creates a new user preset
func (s *Store) CreateUserPreset(ctx context.Context, name, version, description, author, theme, behavior, nodeTypes, relations string) (*UserPresetModel, error) {
	preset := &UserPresetModel{
		PresetID:    uuid.New().String(),
		Name:        name,
		Version:     version,
		Description: description,
		Author:      author,
		Theme:       theme,
		Behavior:    behavior,
		NodeTypes:   nodeTypes,
		Relations:   relations,
	}

	if err := s.db.WithContext(ctx).Create(preset).Error; err != nil {
		return nil, fmt.Errorf("failed to create preset: %w", err)
	}

	return preset, nil
}

// GetUserPreset retrieves a preset by preset_id
func (s *Store) GetUserPreset(ctx context.Context, presetID string) (*UserPresetModel, error) {
	var preset UserPresetModel
	if err := s.db.WithContext(ctx).Where("preset_id = ?", presetID).First(&preset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("preset not found: %s", presetID)
		}
		return nil, fmt.Errorf("failed to get preset: %w", err)
	}
	return &preset, nil
}

// UpdateUserPreset updates a preset
func (s *Store) UpdateUserPreset(ctx context.Context, presetID string, updates map[string]interface{}) (*UserPresetModel, error) {
	result := s.db.WithContext(ctx).Model(&UserPresetModel{}).Where("preset_id = ?", presetID).Updates(updates)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update preset: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("preset not found: %s", presetID)
	}

	return s.GetUserPreset(ctx, presetID)
}

// DeleteUserPreset soft-deletes a preset
func (s *Store) DeleteUserPreset(ctx context.Context, presetID string) error {
	result := s.db.WithContext(ctx).Where("preset_id = ?", presetID).Delete(&UserPresetModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete preset: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("preset not found: %s", presetID)
	}
	return nil
}

// ListUserPresets lists all user presets
func (s *Store) ListUserPresets(ctx context.Context) ([]UserPresetModel, error) {
	var presets []UserPresetModel
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&presets).Error; err != nil {
		return nil, fmt.Errorf("failed to list presets: %w", err)
	}
	return presets, nil
}

// ============================================================================
// Share Link Operations
// ============================================================================

// CreateShareLink creates a new share link for a graph
func (s *Store) CreateShareLink(ctx context.Context, graphID string, password string, expiresIn *time.Duration) (*ShareLinkModel, error) {
	linkID, err := generateLinkID(8)
	if err != nil {
		return nil, fmt.Errorf("failed to generate link ID: %w", err)
	}

	link := &ShareLinkModel{
		LinkID:  linkID,
		GraphID: graphID,
	}

	if password != "" {
		// In production, hash the password properly
		link.Password = password
	}

	if expiresIn != nil {
		expiresAt := time.Now().Add(*expiresIn)
		link.ExpiresAt = &expiresAt
	}

	if err := s.db.WithContext(ctx).Create(link).Error; err != nil {
		return nil, fmt.Errorf("failed to create share link: %w", err)
	}

	return link, nil
}

// GetShareLink retrieves a share link by link_id
func (s *Store) GetShareLink(ctx context.Context, linkID string) (*ShareLinkModel, error) {
	var link ShareLinkModel
	if err := s.db.WithContext(ctx).Where("link_id = ?", linkID).First(&link).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("share link not found: %s", linkID)
		}
		return nil, fmt.Errorf("failed to get share link: %w", err)
	}

	// Check expiration
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return nil, fmt.Errorf("share link expired")
	}

	return &link, nil
}

// GetGraphByShareLink retrieves a graph via share link
func (s *Store) GetGraphByShareLink(ctx context.Context, linkID, password string) (*GraphModel, error) {
	link, err := s.GetShareLink(ctx, linkID)
	if err != nil {
		return nil, err
	}

	// Validate password if set
	if link.Password != "" && link.Password != password {
		return nil, fmt.Errorf("invalid password")
	}

	// Increment view count
	s.db.WithContext(ctx).Model(link).Update("view_count", gorm.Expr("view_count + 1"))

	return s.GetGraph(ctx, link.GraphID)
}

// IncrementViewCount increments the view count for a share link
func (s *Store) IncrementViewCount(ctx context.Context, linkID string) error {
	return s.db.WithContext(ctx).Model(&ShareLinkModel{}).Where("link_id = ?", linkID).Update("view_count", gorm.Expr("view_count + 1")).Error
}

// DeleteShareLink deletes a share link
func (s *Store) DeleteShareLink(ctx context.Context, linkID string) error {
	result := s.db.WithContext(ctx).Where("link_id = ?", linkID).Delete(&ShareLinkModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete share link: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("share link not found: %s", linkID)
	}
	return nil
}

// generateLinkID generates a random link ID
func generateLinkID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// Helper to parse JSON
func parseJSON[T any](data string) (T, error) {
	var result T
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return result, err
	}
	return result, nil
}
