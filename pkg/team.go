package pkg

import "fmt"

const (
	MiaroTeam = 3
	TeamCount = 5
)

// TeamInfo holds display data for the team selector dropdown.
type TeamInfo struct {
	Number   int
	Label    string
	Selected bool
}

// ValidateTeam returns team if it's in [1, TeamCount], otherwise returns MiaroTeam.
func ValidateTeam(team int) int {
	if team < 1 || team > TeamCount {
		return MiaroTeam
	}
	return team
}

// TeamLabel returns the display label for a team number.
func TeamLabel(team int) string {
	if team == MiaroTeam {
		return fmt.Sprintf("Équipe %d (Miaro)", team)
	}
	return fmt.Sprintf("Équipe %d", team)
}

// BuildTeamList returns a slice of TeamInfo for rendering the dropdown.
func BuildTeamList(selectedTeam int) []TeamInfo {
	teams := make([]TeamInfo, TeamCount)
	for i := 0; i < TeamCount; i++ {
		num := i + 1
		teams[i] = TeamInfo{
			Number:   num,
			Label:    TeamLabel(num),
			Selected: num == selectedTeam,
		}
	}
	return teams
}
