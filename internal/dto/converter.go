package dto

import (
	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/samber/lo"
)

//go:generate go tool oapi-codegen --config=../dto/dto.gen.yaml ../../api/swagger.yaml

func (u *User) FillFromEntity(e entity.User) {
	u.IsActive = e.IsActive
	u.TeamName = e.Team.Name
	u.UserId = e.ID
	u.Username = e.Name
}

func (m *TeamMember) ToEntity() *entity.User {
	return &entity.User{
		ID:       m.UserId,
		Name:     m.Username,
		IsActive: m.IsActive,
	}
}

func (t *Team) FillFromEntity(e entity.Team) {
	t.TeamName = e.Name
	t.Members = lo.Map(e.Members, func(u entity.User, _ int) TeamMember {
		return TeamMember{
			IsActive: u.IsActive,
			UserId:   u.ID,
			Username: u.Name,
		}
	})
}

func (t *Team) ToEntity() *entity.Team {
	return &entity.Team{
		Name: t.TeamName,
		Members: lo.Map(t.Members, func(m TeamMember, _ int) entity.User {
			return *m.ToEntity()
		}),
	}
}

func (pr *PullRequest) FillFromEntity(e entity.PullRequest) {
	pr.PullRequestId = e.ID
	pr.PullRequestName = e.Title
	pr.AuthorId = e.AuthorID
	pr.Status = PullRequestStatus(e.Status.Name)
	pr.AssignedReviewers = e.Reviewers
	pr.CreatedAt = &e.CreatedAt
	pr.MergedAt = &e.MergedAt
}
