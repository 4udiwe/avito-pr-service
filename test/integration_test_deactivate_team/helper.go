package integration_test_deactivate_team

import (
	"fmt"
	"net/http"
	"time"

	"github.com/4udiwe/avito-pr-service/internal/api/http/post_pr"
	"github.com/4udiwe/avito-pr-service/internal/api/http/post_team"
	"github.com/4udiwe/avito-pr-service/internal/dto"
	. "github.com/Eun/go-hit"
)

const (
	defaultAttempts = 20
	host            = "app:8080"
	healthPath      = "http://" + host + "/health"
	basePath        = "http://" + host
)

// CreateTeam /team/add
func CreateTeam(teamName string, users []dto.TeamMember) error {
	body := post_team.Request{
		TeamName: teamName,
		Members:  users,
	}

	err := Do(
		Post(basePath+"/team/add"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(body),
		Expect().Status().Equal(http.StatusCreated),
	)
	return err
}

// CreatePR /pullRequest/create
func CreatePR(prID, title, authorID string) error {
	body := post_pr.Request{
		AuthorId:        authorID,
		PullRequestId:   prID,
		PullRequestName: title,
	}

	err := Do(
		Post(basePath+"/pullRequest/create"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(body),
		Expect().Status().Equal(http.StatusCreated),
	)
	return err
}

// HealthCheck
func HealthCheck(attempts int) error {
	var err error
	for attempts > 0 {
		err = Do(Get(healthPath), Expect().Status().Equal(http.StatusOK))
		if err == nil {
			return nil
		}
		time.Sleep(time.Second)
		attempts--
	}
	return err
}

// GenerateUserList creates users for the team given
func GenerateUserList(teamIndex, amount int) []dto.TeamMember {
	users := make([]dto.TeamMember, amount)
	for i := 1; i <= amount; i++ {
		users[i-1] = dto.TeamMember{
			UserId:   fmt.Sprintf("u%d-%d", teamIndex, i),
			Username: fmt.Sprintf("User %d-%d", teamIndex, i),
			IsActive: true,
		}
	}
	return users
}

// GeneratePRID
func GeneratePRID(userID string) string {
	return fmt.Sprintf("pr-%s", userID)
}
