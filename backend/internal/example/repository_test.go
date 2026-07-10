package example

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUnitNormalizePagination_DefaultsAndClamps verifies that page/per_page
// values are normalized to sane defaults and bounds before being turned into
// SQL LIMIT/OFFSET values, matching docs/api-design-standards.md §9
// (page default 1, per_page default 20, max 100).
func TestUnitNormalizePagination_DefaultsAndClamps(t *testing.T) {
	tests := []struct {
		name          string
		page, perPage int
		wantPage      int
		wantPerPage   int
		wantLimit     int
		wantOffset    int
	}{
		{"zero values use defaults", 0, 0, 1, 20, 20, 0},
		{"negative values use defaults", -5, -5, 1, 20, 20, 0},
		{"explicit first page", 1, 10, 1, 10, 10, 0},
		{"explicit second page", 2, 10, 2, 10, 10, 10},
		{"per_page above max clamps to 100", 1, 500, 1, 100, 100, 0},
		{"third page with custom per_page", 3, 25, 3, 25, 25, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, perPage, limit, offset := normalizePagination(tt.page, tt.perPage)

			assert.Equal(t, tt.wantPage, page)
			assert.Equal(t, tt.wantPerPage, perPage)
			assert.Equal(t, tt.wantLimit, limit)
			assert.Equal(t, tt.wantOffset, offset)
		})
	}
}

// TestUnitTotalPages_ComputesCeilingDivision verifies the pagination
// envelope's total_pages is a ceiling division of total/per_page, and that
// zero results yield zero pages (not one empty page).
func TestUnitTotalPages_ComputesCeilingDivision(t *testing.T) {
	tests := []struct {
		name    string
		total   int
		perPage int
		want    int
	}{
		{"no results", 0, 20, 0},
		{"exact multiple", 40, 20, 2},
		{"remainder rounds up", 41, 20, 3},
		{"fewer than one page", 5, 20, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, totalPages(tt.total, tt.perPage))
		})
	}
}
