package repo_pr

import (
	"context"
	"errors"
	"time"

	"github.com/4udiwe/avito-pr-service/internal/entity"
	"github.com/4udiwe/avito-pr-service/internal/repository"
	"github.com/4udiwe/avito-pr-service/pkg/postgres"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"
)

type Repository struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) *Repository {
	return &Repository{pg}
}

func (r *Repository) Create(ctx context.Context, ID, title, authorID string) (entity.PullRequest, error) {
	logrus.Infof("PRRepository.Create: creating PR with title %s", title)

	query, args, _ := r.Builder.Insert("pr").
		Columns("id", "title", "author_id").
		Values(ID, title, authorID).
		Suffix("RETURNING status_id, need_more_reviewers, created_at, merged_at").
		ToSql()

	pr := entity.PullRequest{
		ID:       ID,
		Title:    title,
		AuthorID: authorID,
	}

	err := r.GetTxManager(ctx).QueryRow(ctx, query, args...).Scan(
		&pr.StatusID,
		&pr.NeedMoreReviewers,
		&pr.CreatedAt,
		&pr.MergedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				logrus.Warnf("PRRepository.Create: PR already exists: %s", title)
				return entity.PullRequest{}, repository.ErrPRAlreadyExists
			case pgerrcode.ForeignKeyViolation:
				logrus.Warnf("PRRepository.Create: author not found for PR %s", title)
				return entity.PullRequest{}, repository.ErrAuthorNotFound
			}
		}
		logrus.Errorf("PRRepository.Create: failed to create PR: %v", err)
		return entity.PullRequest{}, err
	}
	return pr, nil
}

func (r *Repository) AssignReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	logrus.Infof("PRRepository.AssignReviewers: assigning reviewers to PR %s", prID)

	queryBuilder := r.Builder.Insert("pr_reviewer").
		Columns("pr_id", "reviewer_id")

	for _, reviewerID := range reviewerIDs {
		queryBuilder = queryBuilder.Values(
			prID,
			reviewerID,
		)
	}
	query, args, _ := queryBuilder.ToSql()

	_, err := r.GetTxManager(ctx).Exec(ctx, query, args...)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				logrus.Warnf("PRRepository.AssignReviewers: reviewer already assigned to PR %s", prID)
				return repository.ErrReviewerAlreadyAssigned
			case pgerrcode.ForeignKeyViolation:
				logrus.Warnf("PRRepository.AssignReviewers: reviewer not found for PR %s", prID)
				return repository.ErrReviewerNotFound
			}
		}
		logrus.Errorf("PRRepository.AssignReviewers: failed to assign reviewers to PR: %v", err)
		return err
	}

	logrus.Infof("PRRepository.AssignReviewers: reviewers assigned to PR %s", prID)
	return nil
}

func (r *Repository) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	logrus.Infof("PRRepository.ReassignReviewer: reassigning reviewer for PR %s", prID)

	query, args, _ := r.Builder.Update("pr_reviewer").
		Set("reviewer_id", newReviewerID).
		Where("pr_id = ? AND reviewer_id = ?", prID, oldReviewerID).
		ToSql()

	_, err := r.GetTxManager(ctx).Exec(ctx, query, args...)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.ForeignKeyViolation {
				logrus.Warnf("PRRepository.ReassignReviewer: new reviewer not found for PR %s", prID)
				return repository.ErrReviewerNotFound
			}
		}
		logrus.Errorf("PRRepository.ReassignReviewer: failed to reassign reviewer for PR: %v", err)
		return err
	}

	logrus.Infof("PRRepository.ReassignReviewer: reviewer reassigned for PR %s", prID)
	return nil
}

func (r *Repository) GetByID(ctx context.Context, ID string) (entity.PullRequest, error) {
	logrus.Infof("PRRepository.GetByID: getting PR by ID %s", ID)

	query, args, _ := r.Builder.Select(
		"id", "title", "author_id", "status_id", "need_more_reviewers", "created_at", "merged_at",
	).From("pr").
		Where("id = ?", ID).
		ToSql()

	var pr entity.PullRequest
	err := r.GetTxManager(ctx).QueryRow(ctx, query, args...).Scan(
		&pr.ID,
		&pr.Title,
		&pr.AuthorID,
		&pr.StatusID,
		&pr.NeedMoreReviewers,
		&pr.CreatedAt,
		&pr.MergedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			logrus.Warnf("PRRepository.GetByID: no PR with ID %s", ID)
			return entity.PullRequest{}, repository.ErrPRNotFound
		}
		logrus.Errorf("PRRepository.GetByID: failed to get PR by ID %s: %v", ID, err)
		return entity.PullRequest{}, err
	}

	logrus.Infof("PRRepository.GetByID: PR found with ID %s", pr.ID)
	return pr, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, ID string, statusID int, mergedAt time.Time) error {
	logrus.Infof("PRRepository.UpdateStatus: updating status for PR %s", ID)

	query, args, _ := r.Builder.Update("pr").Set("status_id", statusID).Where("id = ?", ID).ToSql()

	cmdTag, err := r.GetTxManager(ctx).Exec(ctx, query, args...)

	if err != nil {
		logrus.Errorf("PRRepository.UpdateStatus: failed to update status for PR %s: %v", ID, err)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		logrus.Warnf("PRRepository.UpdateStatus: no PR with ID %s to update", ID)
		return repository.ErrPRNotFound
	}

	logrus.Infof("PRRepository.UpdateStatus: status updated for PR %s", ID)
	return nil
}

func (r *Repository) GetReviewersByPR(ctx context.Context, prID string) ([]entity.PRReviewer, error) {
	query, args, _ := r.Builder.Select(
		"id", "pr_id", "reviewer_id", "assigned_at",
	).From("pr_reviewer").
		Where("pr_id = ?", prID).
		ToSql()

	rows, err := r.GetTxManager(ctx).Query(ctx, query, args...)
	if err != nil {
		logrus.Errorf("PRRepository.GetReviewersByPR: failed to get reviewers for PR %s: %v", prID, err)
		return nil, err
	}
	defer rows.Close()

	var reviewers []entity.PRReviewer
	for rows.Next() {
		var rev entity.PRReviewer
		if err := rows.Scan(&rev.ID, &rev.PRID, &rev.ReviewerID, &rev.AssignedAt); err != nil {
			logrus.Errorf("PRRepository.GetReviewersByPR: failed to scan reviewer for PR %s: %v", prID, err)
			return nil, err
		}
		reviewers = append(reviewers, rev)
	}

	logrus.Infof("PRRepository.GetReviewersByPR: reviewers found for PR %s", prID)
	return reviewers, nil
}

func (r *Repository) ListByReviewer(ctx context.Context, reviewerID string) ([]entity.PullRequest, error) {
	query, args, _ := r.Builder.Select(
		"p.id", "p.title", "p.author_id", "p.status_id", "p.need_more_reviewers", "p.created_at", "p.merged_at",
	).From("pr AS p").
		Join("pr_reviewer AS r ON p.id = r.pr_id").
		Where("r.reviewer_id = ?", reviewerID).
		ToSql()

	rows, err := r.GetTxManager(ctx).Query(ctx, query, args...)
	if err != nil {
		logrus.Errorf("PRRepository.ListByReviewer: failed to list PRs for reviewer %s: %v", reviewerID, err)
		return nil, err
	}
	defer rows.Close()

	var prs []entity.PullRequest
	for rows.Next() {
		var pr entity.PullRequest
		if err := rows.Scan(
			&pr.ID, &pr.Title, &pr.AuthorID, &pr.StatusID,
			&pr.NeedMoreReviewers, &pr.CreatedAt, &pr.MergedAt,
		); err != nil {
			logrus.Errorf("PRRepository.ListByReviewer: failed to scan PR for reviewer %s: %v", reviewerID, err)
			return nil, err
		}
		prs = append(prs, pr)
	}

	logrus.Infof("PRRepository.ListByReviewer: PRs found for reviewer %s", reviewerID)
	return prs, nil
}

func (r *Repository) GetPRStatuses(ctx context.Context) ([]entity.PRStatus, error) {
	logrus.Infof("PRRepository.GetPRStatuses: getting all PR statuses")

	query, args, _ := r.Builder.Select(
		"id", "name",
	).From("pr_status").
		ToSql()

	var statuses []entity.PRStatus

	rows, err := r.GetTxManager(ctx).Query(ctx, query, args...)
	if err != nil {
		logrus.Errorf("PRRepository.GetPRStatuses: failed to get PR statuses: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status entity.PRStatus
		if err := rows.Scan(&status.ID, &status.Name); err != nil {
			logrus.Errorf("PRRepository.GetPRStatuses: failed to scan PR status: %v", err)
			return nil, err
		}
		statuses = append(statuses, status)
	}

	logrus.Infof("PRRepository.GetPRStatuses: PR statuses retrieved")
	return statuses, nil
}
