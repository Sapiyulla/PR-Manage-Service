package repository

import (
	"context"
	"errors"
	"pr-manage-service/internal/domain"
	"pr-manage-service/pkg/errs"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type PullRequestRepository struct {
	ctx      context.Context
	pool     *pgxpool.Pool
	rtimeout time.Duration
}

// GetWithUser implements domain.PRRepository.
func (r *PullRequestRepository) GetWithUser(user *domain.User) (*[]domain.PullRequest, error) {
	reqCtx, cancel := context.WithTimeout(r.ctx, r.rtimeout)
	defer cancel()

	tx, err := r.pool.Begin(reqCtx)
	if err != nil {
		logrus.Error(logPrefix, err.Error())
		return nil, &errs.InternalError{}
	}
	defer tx.Rollback(reqCtx)

	var internalUserID int
	err = tx.QueryRow(reqCtx,
		`SELECT id FROM users WHERE user_id = $1 AND team_name = $2`,
		user.UserID, user.TeamName).Scan(&internalUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &errs.NotFoundError{Domain: "user"}
		}
		logrus.Error(logPrefix, err.Error())
		return nil, &errs.InternalError{}
	}

	rows, err := tx.Query(reqCtx, `
        SELECT 
            p.id, p.name, 
            author.user_id as author_user_id, 
            author.team_name as author_team_name,
            p.status, p.need_more_reviewers, p.created_at, p.updated_at
        FROM prs p
        JOIN users author ON p.author_id = author.id
        WHERE p.author_id = $1 
           OR p.id IN (
               SELECT pr_id FROM pr_reviewers WHERE user_id = $1
           )
        ORDER BY p.created_at DESC
    `, internalUserID)

	if err != nil {
		logrus.Error(logPrefix, err.Error())
		return nil, &errs.InternalError{}
	}
	defer rows.Close()

	var pullRequests []domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		var status string

		err := rows.Scan(
			&pr.PrID, &pr.PrName,
			&pr.AuthorID, &pr.TeamName,
			&status, &pr.NeedMoreReviewers, &pr.CreatedAt, &pr.UpdatedAt,
		)
		if err != nil {
			logrus.Error(logPrefix, err.Error())
			return nil, &errs.InternalError{}
		}

		// Конвертируем строковый статус в доменный тип
		if status == string(domain.OPEN) {
			pr.Status = domain.OPEN
		} else if status == "MERGED" {
			pr.Status = domain.MERGED
		}

		pullRequests = append(pullRequests, pr)
	}

	if err := rows.Err(); err != nil {
		logrus.Error(logPrefix, err.Error())
		return nil, &errs.InternalError{}
	}

	if err := tx.Commit(reqCtx); err != nil {
		logrus.Error(logPrefix, err.Error())
		return nil, &errs.InternalError{}
	}

	return &pullRequests, nil
}

func NewPullRequestRepository(ctx context.Context, pool *pgxpool.Pool, rtimeout time.Duration) domain.PRRepository {
	return &PullRequestRepository{
		ctx:      ctx,
		pool:     pool,
		rtimeout: rtimeout,
	}
}

// CreateNewPR implements domain.PRRepository.
func (r *PullRequestRepository) CreateNewPR(pr *domain.PullRequest) (assigned_reviewers []string, err error) {
	reqCtx, cancel := context.WithTimeout(r.ctx, r.rtimeout)
	defer cancel()

	tx, err := r.pool.Begin(reqCtx)
	if err != nil {
		logrus.Error(logPrefix, err.Error())
		return nil, &errs.InternalError{}
	}
	defer tx.Rollback(reqCtx)

	// Получаем internal author_id (serial) по user_id (varchar)
	var authorInternalID int
	if err := tx.QueryRow(reqCtx,
		`SELECT id FROM users WHERE user_id=$1 AND team_name=$2 AND is_active=true`,
		pr.AuthorID, pr.TeamName).Scan(&authorInternalID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &errs.NotFoundError{
				Domain: "resource",
			}
		}
		logrus.Error(logPrefix, err.Error())
		return nil, &errs.InternalError{}
	}

	// Находим доступных ревьюверов из той же команды (исключая автора)
	rows, err := tx.Query(reqCtx,
		`SELECT user_id, id FROM users 
         WHERE team_name = $1 AND is_active = true AND user_id != $2 
         ORDER BY id LIMIT 2`,
		pr.TeamName, pr.AuthorID)
	if err != nil {
		logrus.Error(logPrefix, "Failed to find reviewers: "+err.Error())
		return nil, &errs.InternalError{}
	}
	defer rows.Close()

	var reviewerUserIDs []string
	var reviewerInternalIDs []int
	for rows.Next() {
		var userID string
		var internalID int
		if err := rows.Scan(&userID, &internalID); err != nil {
			logrus.Error(logPrefix, err.Error())
			return nil, &errs.InternalError{}
		}
		reviewerUserIDs = append(reviewerUserIDs, userID)
		reviewerInternalIDs = append(reviewerInternalIDs, internalID)
	}

	// Определяем статус need_more_reviewers
	needMoreReviewers := len(reviewerInternalIDs) < 2

	// Устанавливаем поля PR до вставки
	now := time.Now()
	pr.Status = domain.OPEN
	pr.NeedMoreReviewers = needMoreReviewers
	pr.CreatedAt = now
	pr.UpdatedAt = now

	// Вставляем PR с правильным author_id (internal ID)
	if _, err := tx.Exec(reqCtx,
		`INSERT INTO prs (id, name, author_id, status, need_more_reviewers, created_at, updated_at) 
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		pr.PrID, pr.PrName, authorInternalID, pr.Status, pr.NeedMoreReviewers, pr.CreatedAt, pr.UpdatedAt); err != nil {
		if strings.Contains(err.Error(), "dublicate") || strings.Contains(err.Error(), "duplicate") {
			return nil, &errs.AlreadyExistsError{
				Domain: "pr '" + pr.PrID + "'",
			}
		}
		logrus.Error(logPrefix, err.Error())
		return nil, &errs.InternalError{}
	}

	// Назначаем найденных ревьюверов
	for _, reviewerInternalID := range reviewerInternalIDs {
		if _, err := tx.Exec(reqCtx,
			`INSERT INTO pr_reviewers (pr_id, user_id, team_name) VALUES ($1, $2, $3)`,
			pr.PrID, reviewerInternalID, pr.TeamName); err != nil {
			logrus.Error(logPrefix, err.Error())
			return nil, &errs.InternalError{}
		}
	}

	if err := tx.Commit(reqCtx); err != nil {
		logrus.Error(logPrefix, err.Error())
		return nil, &errs.InternalError{}
	}

	return reviewerUserIDs, nil
}

// Merge implements domain.PRRepository.
func (r *PullRequestRepository) Merge(prID string) (pr *domain.PullRequest, assigned_reviewers []string, err error) {
	reqCtx, cancel := context.WithTimeout(r.ctx, r.rtimeout)
	defer cancel()

	tx, err := r.pool.Begin(reqCtx)
	if err != nil {
		logrus.Error(logPrefix, err.Error())
		return nil, nil, &errs.InternalError{}
	}
	defer tx.Rollback(reqCtx)

	// Получаем текущий статус PR и информацию об авторе
	var (
		status            string
		prName            string
		authorInternalID  int
		needMoreReviewers bool
		updatedAt         time.Time
	)

	pr = &domain.PullRequest{}

	err = tx.QueryRow(reqCtx, `
        SELECT p.status, p.name, p.author_id, p.need_more_reviewers, p.updated_at,
               u.user_id, u.team_name
        FROM prs p
        JOIN users u ON p.author_id = u.id
        WHERE p.id = $1
    `, prID).Scan(&status, &prName, &authorInternalID, &needMoreReviewers, &updatedAt,
		&pr.AuthorID, &pr.TeamName)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, &errs.NotFoundError{Domain: "pull request"}
		}
		logrus.Error(logPrefix, err.Error())
		return nil, nil, &errs.InternalError{}
	}

	// Инициализируем объект PR
	pr = &domain.PullRequest{
		PrID:              prID,
		PrName:            prName,
		AuthorID:          pr.AuthorID,
		TeamName:          pr.TeamName,
		NeedMoreReviewers: needMoreReviewers,
		UpdatedAt:         updatedAt,
	}

	// Преобразуем статус
	if status == "MERGED" {
		pr.Status = domain.MERGED
		// Если уже мержжен, просто возвращаем информацию без изменений
	} else {
		pr.Status = domain.OPEN // временно, обновим ниже
	}

	// Получаем список назначенных ревьюверов
	rows, err := tx.Query(reqCtx, `
        SELECT u.user_id 
        FROM pr_reviewers prr
        JOIN users u ON prr.user_id = u.id
        WHERE prr.pr_id = $1
    `, prID)

	if err != nil {
		logrus.Error(logPrefix, err.Error())
		return nil, nil, &errs.InternalError{}
	}
	defer rows.Close()

	for rows.Next() {
		var reviewerUserID string
		if err := rows.Scan(&reviewerUserID); err != nil {
			logrus.Error(logPrefix, err.Error())
			return nil, nil, &errs.InternalError{}
		}
		assigned_reviewers = append(assigned_reviewers, reviewerUserID)
	}

	// Если PR уже мержжен, просто возвращаем данные
	if status == "MERGED" {
		if err := tx.Commit(reqCtx); err != nil {
			logrus.Error(logPrefix, err.Error())
			return nil, nil, &errs.InternalError{}
		}
		return pr, assigned_reviewers, nil
	}

	// Обновляем статус на MERGED
	now := time.Now()
	if _, err := tx.Exec(reqCtx, `
        UPDATE prs 
        SET status = 'MERGED', updated_at = $1 
        WHERE id = $2
    `, now, prID); err != nil {
		logrus.Error(logPrefix, err.Error())
		return nil, nil, &errs.InternalError{}
	}

	// Обновляем объект PR
	pr.Status = domain.MERGED
	pr.UpdatedAt = now

	if err := tx.Commit(reqCtx); err != nil {
		logrus.Error(logPrefix, err.Error())
		return nil, nil, &errs.InternalError{}
	}

	return pr, assigned_reviewers, nil
}
