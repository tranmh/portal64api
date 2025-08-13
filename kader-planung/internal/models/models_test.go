package models

import "testing"

func TestCalculateClubIDPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		clubID   string
		expected1 string
		expected2 string
		expected3 string
	}{
		{
			name:     "Normal club ID",
			clubID:   "C0327",
			expected1: "C",
			expected2: "C0",
			expected3: "C03",
		},
		{
			name:     "Short club ID (2 chars)",
			clubID:   "B5", 
			expected1: "B",
			expected2: "B5",
			expected3: "",
		},
		{
			name:     "Single character club ID",
			clubID:   "A",
			expected1: "A",
			expected2: "",
			expected3: "",
		},
		{
			name:     "Empty club ID",
			clubID:   "",
			expected1: "",
			expected2: "",
			expected3: "",
		},
		{
			name:     "Long club ID",
			clubID:   "C0327123",
			expected1: "C",
			expected2: "C0",
			expected3: "C03",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix1, prefix2, prefix3 := CalculateClubIDPrefixes(tt.clubID)
			
			if prefix1 != tt.expected1 {
				t.Errorf("CalculateClubIDPrefixes(%q) prefix1 = %q, expected %q", tt.clubID, prefix1, tt.expected1)
			}
			if prefix2 != tt.expected2 {
				t.Errorf("CalculateClubIDPrefixes(%q) prefix2 = %q, expected %q", tt.clubID, prefix2, tt.expected2)
			}
			if prefix3 != tt.expected3 {
				t.Errorf("CalculateClubIDPrefixes(%q) prefix3 = %q, expected %q", tt.clubID, prefix3, tt.expected3)
			}
		})
	}
}
