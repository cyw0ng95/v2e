package testutils

import (
"testing"

"gorm.io/gorm"
)

func TestLevelDemo(t *testing.T) {
// Level 1 test - always runs
Run(t, Level1, "Level1Only", nil, func(t *testing.T, tx *gorm.DB) {
t.Log("Level 1 test executed")
})

// Level 2 test - runs at Level 2 and 3
Run(t, Level2, "Level2Only", nil, func(t *testing.T, tx *gorm.DB) {
t.Log("Level 2 test executed")
})

// Level 3 test - runs only at Level 3
Run(t, Level3, "Level3Only", nil, func(t *testing.T, tx *gorm.DB) {
t.Log("Level 3 test executed")
})
}
