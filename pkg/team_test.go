package pkg

import (
	"testing"
)

func TestValidateTeam(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"valid team 1", 1, 1},
		{"valid team 5", 5, 5},
		{"zero defaults to 1", 0, 1},
		{"negative defaults to 1", -1, 1},
		{"too high defaults to 1", 6, 1},
		{"way too high defaults to 1", 100, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateTeam(tc.input)
			if result != tc.expected {
				t.Errorf("ValidateTeam(%d) = %d, want %d", tc.input, result, tc.expected)
			}
		})
	}
}

func TestTeamLabel(t *testing.T) {
	tests := []struct {
		team     int
		expected string
	}{
		{1, "Équipe 1 (Miaro)"},
		{2, "Équipe 2"},
		{3, "Équipe 3"},
		{4, "Équipe 4"},
		{5, "Équipe 5"},
	}
	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := TeamLabel(tc.team)
			if result != tc.expected {
				t.Errorf("TeamLabel(%d) = %q, want %q", tc.team, result, tc.expected)
			}
		})
	}
}

func TestBuildTeamList(t *testing.T) {
	teams := BuildTeamList(3)
	if len(teams) != TeamCount {
		t.Fatalf("expected %d teams, got %d", TeamCount, len(teams))
	}
	if !teams[2].Selected {
		t.Error("team 3 should be selected")
	}
	if teams[0].Selected {
		t.Error("team 1 should not be selected")
	}
	if teams[0].Label != "Équipe 1 (Miaro)" {
		t.Errorf("team 1 label = %q, want 'Équipe 1 (Miaro)'", teams[0].Label)
	}
}
