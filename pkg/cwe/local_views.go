package cwe

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GORM models for views and nested arrays
type ViewModel struct {
	ID        string `gorm:"column:id;primaryKey"`
	Name      string `gorm:"column:name"`
	Type      string `gorm:"column:type"`
	Status    string `gorm:"column:status"`
	Objective string `gorm:"column:objective"`
	Raw       []byte `gorm:"column:raw;type:blob"`
}

type ViewMemberModel struct {
	ID     uint   `gorm:"column:id;primaryKey"`
	ViewID string `gorm:"column:view_id;index"` // Foreign key to parent view
	CWEID  string `gorm:"column:cwe_id"`        // Foreign key to CWE
	Role   string `gorm:"column:role"`
}

type StakeholderModel struct {
	ID          uint   `gorm:"column:id;primaryKey"`
	ViewID      string `gorm:"column:view_id;index"` // Foreign key to parent view
	Type        string `gorm:"column:type"`
	Description string `gorm:"column:description"`
}

type ViewReferenceModel struct {
	ID                  uint   `gorm:"column:id;primaryKey"`
	ViewID              string `gorm:"column:view_id;index"` // Foreign key to parent view
	ExternalReferenceID string `gorm:"column:external_reference_id"`
	Title               string `gorm:"column:title"`
	URL                 string `gorm:"column:url"`
	Description         string `gorm:"column:description"`
}

type ViewNoteModel struct {
	ID     uint   `gorm:"column:id;primaryKey"`
	ViewID string `gorm:"column:view_id;index"` // Foreign key to parent view
	Type   string `gorm:"column:type"`
	Note   string `gorm:"column:note"`
}

type ViewContentHistoryModel struct {
	ID             uint   `gorm:"column:id;primaryKey"`
	ViewID         string `gorm:"column:view_id;index"` // Foreign key to parent view
	Type           string `gorm:"column:type"`
	SubmissionName string `gorm:"column:submission_name"`
	Date           string `gorm:"column:date"`
	Version        string `gorm:"column:version"`
	Details        string `gorm:"column:details"`
}

// AutoMigrateViews migrates view-related tables into the provided DB.
func AutoMigrateViews(db *gorm.DB) error {
	return db.AutoMigrate(
		&ViewModel{},
		&ViewMemberModel{},
		&StakeholderModel{},
		&ViewReferenceModel{},
		&ViewNoteModel{},
		&ViewContentHistoryModel{},
	)
}

// SaveView inserts or updates a CWE view and its nested arrays in a transaction.
func (s *LocalCWEStore) SaveView(ctx context.Context, v *CWEView) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	vm := ViewModel{
		ID:        v.ID,
		Name:      v.Name,
		Type:      v.Type,
		Status:    v.Status,
		Objective: v.Objective,
		Raw:       v.Raw,
	}
	if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&vm).Error; err != nil {
		tx.Rollback()
		return err
	}
	// members
	if err := tx.Where("view_id = ?", v.ID).Delete(&ViewMemberModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, m := range v.Members {
		mm := ViewMemberModel{ViewID: v.ID, CWEID: m.CWEID, Role: m.Role}
		if err := tx.Create(&mm).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	// stakeholders
	if err := tx.Where("view_id = ?", v.ID).Delete(&StakeholderModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, a := range v.Audience {
		am := StakeholderModel{ViewID: v.ID, Type: a.Type, Description: a.Description}
		if err := tx.Create(&am).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	// references
	if err := tx.Where("view_id = ?", v.ID).Delete(&ViewReferenceModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, r := range v.References {
		rm := ViewReferenceModel{
			ViewID:              v.ID,
			ExternalReferenceID: r.ExternalReferenceID,
			Title:               r.Title,
			URL:                 r.URL,
			Description:         r.Section,
		}
		if err := tx.Create(&rm).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	// notes
	if err := tx.Where("view_id = ?", v.ID).Delete(&ViewNoteModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, n := range v.Notes {
		nm := ViewNoteModel{ViewID: v.ID, Type: n.Type, Note: n.Note}
		if err := tx.Create(&nm).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	// content history (simplified)
	if err := tx.Where("view_id = ?", v.ID).Delete(&ViewContentHistoryModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, ch := range v.ContentHistory {
		chm := ViewContentHistoryModel{
			ViewID:         v.ID,
			Type:           ch.Type,
			SubmissionName: ch.SubmissionName,
			Date:           ch.Date,
			Version:        ch.Version,
			Details:        ch.SubmissionComment + ch.ModificationComment + ch.ContributionComment,
		}
		if err := tx.Create(&chm).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// GetViewByID loads a view and its nested arrays.
func (s *LocalCWEStore) GetViewByID(ctx context.Context, id string) (*CWEView, error) {
	var m ViewModel
	if err := s.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	v := &CWEView{ID: m.ID, Name: m.Name, Type: m.Type, Status: m.Status, Objective: m.Objective, Raw: m.Raw}
	var members []ViewMemberModel
	s.db.WithContext(ctx).Where("view_id = ?", id).Find(&members)
	for _, mm := range members {
		v.Members = append(v.Members, ViewMember{CWEID: mm.CWEID, Role: mm.Role})
	}
	var aud []StakeholderModel
	s.db.WithContext(ctx).Where("view_id = ?", id).Find(&aud)
	for _, a := range aud {
		v.Audience = append(v.Audience, Stakeholder{Type: a.Type, Description: a.Description})
	}
	var refs []ViewReferenceModel
	s.db.WithContext(ctx).Where("view_id = ?", id).Find(&refs)
	for _, r := range refs {
		v.References = append(v.References, Reference{
			ExternalReferenceID: r.ExternalReferenceID,
			Title:               r.Title,
			URL:                 r.URL,
			Section:             r.Description,
		})
	}
	var notes []ViewNoteModel
	s.db.WithContext(ctx).Where("view_id = ?", id).Find(&notes)
	for _, n := range notes {
		v.Notes = append(v.Notes, Note{Type: n.Type, Note: n.Note})
	}
	var chs []ViewContentHistoryModel
	s.db.WithContext(ctx).Where("view_id = ?", id).Find(&chs)
	for _, ch := range chs {
		v.ContentHistory = append(v.ContentHistory, ContentHistory{
			Type:           ch.Type,
			SubmissionName: ch.SubmissionName,
			Date:           ch.Date,
			Version:        ch.Version,
		})
	}
	return v, nil
}

// ListViewsPaginated returns paginated views.
func (s *LocalCWEStore) ListViewsPaginated(ctx context.Context, offset, limit int) ([]CWEView, int64, error) {
	var models []ViewModel
	var total int64
	s.db.Model(&ViewModel{}).Count(&total)
	if err := s.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	out := make([]CWEView, 0, len(models))
	for _, m := range models {
		out = append(out, CWEView{ID: m.ID, Name: m.Name, Type: m.Type, Status: m.Status, Objective: m.Objective, Raw: m.Raw})
	}
	return out, total, nil
}

// DeleteView deletes a view and nested rows.
func (s *LocalCWEStore) DeleteView(ctx context.Context, id string) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Where("id = ?", id).Delete(&ViewModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	tables := []interface{}{&ViewMemberModel{}, &StakeholderModel{}, &ViewReferenceModel{}, &ViewNoteModel{}, &ViewContentHistoryModel{}}
	for _, t := range tables {
		if err := tx.Where("view_id = ?", id).Delete(t).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}
