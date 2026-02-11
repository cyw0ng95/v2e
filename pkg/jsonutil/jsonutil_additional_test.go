package jsonutil

import (
	"fmt"
	"reflect"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

// typedStruct is used for round-trip marshaling tests.
type typedStruct struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Flag  bool   `json:"flag"`
}

func TestMarshalUnmarshal_RoundTripManyValues(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMarshalUnmarshal_RoundTripManyValues", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name  string
			value interface{}
		}{
			{name: "string", value: "hello"},
			{name: "bool-true", value: true},
			{name: "bool-false", value: false},
			{name: "float", value: 3.14},
			{name: "map", value: map[string]int{"a": 1, "b": 2}},
			{name: "slice", value: []string{"x", "y", "z"}},
			{name: "struct", value: typedStruct{Name: "alpha", Count: 2, Flag: true}},
			{name: "empty-map", value: map[string]string{}},
			{name: "empty-slice", value: []int{}},
			{name: "zero-int", value: 0},
		}

		// Add a large number of generated integer and string cases to exercise edge/value handling.
		for i := 0; i < 120; i++ {
			cases = append(cases, struct {
				name  string
				value interface{}
			}{
				name:  fmt.Sprintf("int-%03d", i),
				value: i,
			})
			cases = append(cases, struct {
				name  string
				value interface{}
			}{
				name:  fmt.Sprintf("str-%03d", i),
				value: fmt.Sprintf("val-%03d", i),
			})
		}

		for _, tc := range cases {
			tc := tc // capture
			t.Run(tc.name, func(t *testing.T) {
				data, err := Marshal(tc.value)
				if err != nil {
					t.Fatalf("Marshal failed for %s: %v", tc.name, err)
				}

				targetPtr := reflect.New(reflect.TypeOf(tc.value))
				if err := Unmarshal(data, targetPtr.Interface()); err != nil {
					t.Fatalf("Unmarshal failed for %s: %v", tc.name, err)
				}

				got := reflect.Indirect(targetPtr).Interface()
				if !reflect.DeepEqual(got, tc.value) {
					t.Fatalf("round-trip mismatch for %s: want %v got %v", tc.name, tc.value, got)
				}
			})
		}
	})

}

func TestMarshalIndent_RoundTripSmallMap(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMarshalIndent_RoundTripSmallMap", nil, func(t *testing.T, tx *gorm.DB) {
		original := map[string]string{"alpha": "1", "beta": "2"}
		data, err := MarshalIndent(original, "", "  ")
		if err != nil {
			t.Fatalf("MarshalIndent failed: %v", err)
		}

		var decoded map[string]string
		if err := Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if !reflect.DeepEqual(decoded, original) {
			t.Fatalf("MarshalIndent round-trip mismatch: want %v got %v", original, decoded)
		}
	})

}
