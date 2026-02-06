# Memory FSM Learning System - Data Migration Guide

## Overview

This guide explains the data migration strategy for the Memory FSM Learning System, including the URN field addition and FSM state initialization.

## Migration Requirements

### Phase 1: URN Generation for Existing Data

**Files:**
- `pkg/notes/migration.go` - MigrateExistingData function

**Migration Steps:**

1. **Bookmark URN Generation**
   - Generate URN for each existing bookmark
   - Format: `v2e::<provider>::<type>::<id>`
   - Example: `v2e::nvd::cve::CVE-2021-1234`

2. **Note URN Generation**
   - Generate URN for each existing note
   - Format: `v2e::note::<id>`
   - Example: `v2e::note::123`

3. **Memory Card URN Generation**
   - Generate URN for each existing memory card
   - Format: `v2e::card::<id>`
   - Example: `v2e::card::456`

4. **Bidirectional URN Link Creation**
   - Create URNIndex entries for each URN link
   - Maintain bidirectional relationships
   - Support reverse lookups

### Phase 2: FSM State Initialization

**Default States:**

1. **Notes**
   - New notes: `draft`
   - Existing notes with completion: `learned` (if has valid content)

2. **Memory Cards**
   - New cards: `new`
   - Cards with reviews: `reviewed` (if has review history)
   - Mastered cards: `mastered` (if repetition >= 5)

3. **Bookmarks**
   - All bookmarks: `to-review` (default learning state)

### Phase 3: State History Initialization

**History Generation:**

1. **Initial State History**
   - Create initial state history entry for each object
   - Timestamp: object creation time
   - Reason: "migration"
   - User ID: "system"

2. **State History JSON**
   - Store as JSON in FSMStateHistory column
   - Format: `[{"from_state":"draft","to_state":"learned","timestamp":"2024-01-01T00:00:00Z","reason":"migration","user_id":"system"}]`

## Migration Implementation

### Migration Function

```go
func MigrateExistingData(db *gorm.DB, noteService *NoteService, memoryCardService *MemoryCardService) error {
    // Generate URNs for existing bookmarks
    var bookmarks []BookmarkModel
    if err := db.Find(&bookmarks).Error; err != nil {
        return fmt.Errorf("failed to load bookmarks: %w", err)
    }

    for _, bookmark := range bookmarks {
        if bookmark.URN == "" {
            bookmark.URN = GenerateURN(bookmark.ItemType, bookmark.ItemID, "")
            if err := db.Save(&bookmark).Error; err != nil {
                return fmt.Errorf("failed to update bookmark URN: %w", err)
            }
        }
    }

    // Generate URNs for existing notes
    var notes []NoteModel
    if err := db.Find(&notes).Error; err != nil {
        return fmt.Errorf("failed to load notes: %w", err)
    }

    for _, note := range notes {
        if note.URN == "" {
            note.URN = GenerateURN("note", "", fmt.Sprintf("%d", note.ID))
            if err := db.Save(&note).Error; err != nil {
                return fmt.Errorf("failed to update note URN: %w", err)
            }
        }

        // Initialize FSM state
        if note.FSMState == "" {
            note.FSMState = "draft"
            if err := db.Save(&note).Error; err != nil {
                return fmt.Errorf("failed to initialize note FSM state: %w", err)
            }
        }
    }

    // Generate URNs for existing memory cards
    var cards []MemoryCardModel
    if err := db.Find(&cards).Error; err != nil {
        return fmt.Errorf("failed to load memory cards: %w", err)
    }

    for _, card := range cards {
        if card.URN == "" {
            card.URN = GenerateURN("card", "", fmt.Sprintf("%d", card.ID))
            if err := db.Save(&card).Error; err != nil {
                return fmt.Errorf("failed to update card URN: %w", err)
            }
        }

        // Initialize FSM state based on current status
        if card.FSMState == "" {
            card.FSMState = statusToFSMState(CardStatus(card.Status))
            if err := db.Save(&card).Error; err != nil {
                return fmt.Errorf("failed to initialize card FSM state: %w", err)
            }
        }
    }

    return nil
}
```

### URN Generation Function

```go
func GenerateURN(provider, itemType, itemID string) string {
    if provider == "" {
        return ""
    }

    var parts []string
    parts = append(parts, "v2e")
    parts = append(parts, provider)

    if itemType != "" {
        parts = append(parts, itemType)
    }

    if itemID != "" {
        parts = append(parts, itemID)
    }

    return strings.Join(parts, "::")
}

func GetNoteURN(id uint) string {
    return fmt.Sprintf("v2e::note::%d", id)
}

func GetCardURN(id uint) string {
    return fmt.Sprintf("v2e::card::%d", id)
}
```

### Status to FSM State Mapping

```go
func statusToFSMState(status CardStatus) fsm.MemoryState {
    switch status {
    case StatusNew:
        return fsm.MemoryStateNew
    case StatusLearning:
        return fsm.MemoryStateLearning
    case StatusReviewed:
        return fsm.MemoryStateReviewed
    case StatusMastered:
        return fsm.MemoryStateMastered
    case StatusArchived:
        return fsm.MemoryStateArchived
    default:
        return fsm.MemoryStateNew
    }
}
```

## Backward Compatibility

### Existing Data Handling

1. **Missing URNs**
   - Generate URNs for existing records
   - Maintain relationships by using existing IDs

2. **Missing FSM States**
   - Initialize with appropriate default states
   - Create initial state history entries

3. **Empty Content Fields**
   - Initialize with empty TipTap JSON document
   - Format: `{"type":"doc","content":[{"type":"paragraph","content":[]}]}`

4. **Relationship Integrity**
   - Preserve existing Bookmark → Note relationships
   - Preserve existing Bookmark → MemoryCard relationships
   - Create URNIndex entries for all URN links

### Data Validation

After migration, validate:

1. **URN Uniqueness**
   - All URNs should be unique
   - No duplicate URN values

2. **FSM State Validity**
   - All FSM states should be valid
   - No invalid state values

3. **State History Integrity**
   - All state history should be valid JSON
   - Timestamps should be chronological

4. **Relationship Completeness**
   - All bookmarks should have URN
   - All notes should have URN
   - All memory cards should have URN

## Rollback Plan

If migration fails:

1. **Database Backup**
   - Create full database backup before migration
   - Store backup in safe location

2. **Rollback Steps**
   - Restore from backup
   - Verify data integrity

3. **Error Handling**
   - Log all migration errors
   - Provide detailed error messages
   - Allow partial rollback if needed

## Testing

### Unit Tests

1. **Test URN Generation**
   - Verify URN format
   - Verify uniqueness
   - Test edge cases

2. **Test FSM State Initialization**
   - Verify default states
   - Test state transitions
   - Validate state history

3. **Test Data Migration**
   - Test migration with sample data
   - Verify backward compatibility
   - Test error scenarios

### Integration Tests

1. **Test Complete Migration**
   - Run full migration on test database
   - Verify all records migrated
   - Validate data integrity

2. **Test Service Functionality**
   - Verify services work with migrated data
   - Test CRUD operations
   - Verify URN lookups

## Performance Considerations

1. **Batch Updates**
   - Use batch updates for large datasets
   - Minimize database round trips

2. **Transaction Safety**
   - Use transactions for related updates
   - Rollback on errors

3. **Memory Usage**
   - Process records in batches
   - Avoid loading entire dataset into memory

## Security Considerations

1. **Data Privacy**
   - No user data exposed during migration
   - URNs contain no sensitive information

2. **Access Control**
   - Migration requires database write access
   - Log all migration operations

3. **Audit Trail**
   - Record migration start/end times
   - Track number of records migrated
   - Log any errors or warnings

## Support

For issues or questions about migration:

1. Check logs for detailed error messages
2. Verify database backup exists
3. Review this guide for common issues
4. Contact support team if needed
