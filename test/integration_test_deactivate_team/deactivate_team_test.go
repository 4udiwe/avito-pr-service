//go:build integration

package integration_test_deactivate_team

import (
	"net/http"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
)

func TestDeactivateTeamAndReassignPRs(t *testing.T) {
	const (
		teamsAmount  = 100
		usersPerTeam = 1000
	)

	// Create teams with users
	for teamIdx := 1; teamIdx <= teamsAmount; teamIdx++ {
		teamName := "Team-" + string(rune(teamIdx))
		users := GenerateUserList(teamIdx, usersPerTeam)

		if err := CreateTeam(teamName, users); err != nil {
			t.Fatalf("failed to create team %d: %v", teamIdx, err)
		}
	}
	// For the first team create PR for each user
	firstTeamUsers := GenerateUserList(1, usersPerTeam)
	for _, u := range firstTeamUsers {
		userID := u.UserId
		prID := GeneratePRID(userID)

		if err := CreatePR(prID, "PR for "+userID, userID); err != nil {
			t.Fatalf("failed to create PR for user %s: %v", userID, err)
		}
	}

	// Deactivate first team and measure time it takes
	teamToDeactivate := "Team-1"
	start := time.Now()
	err := Do(
		Post(basePath+"/team/deactivate"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(map[string]string{"team_name": teamToDeactivate}),
		Expect().Status().Equal(http.StatusOK),
	)
	duration := time.Since(start)
	if err != nil {
		t.Fatalf("failed to deactivate team: %v", err)
	}

	t.Logf("DeactivateTeamAndReassignPRs took %s", duration)
}
