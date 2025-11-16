package repository

import (
	"context"
	"errors"
	"pr-manage-service/internal/domain"
	"pr-manage-service/pkg/errs"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type userRepository struct {
	ctx      context.Context
	pool     *pgxpool.Pool
	rtimeout time.Duration
}

func NewUserRepository(ctx context.Context, pool *pgxpool.Pool, rtimeout time.Duration) domain.UserRepository {
	return &userRepository{
		ctx:      ctx,
		pool:     pool,
		rtimeout: rtimeout,
	}
}

// ChangeActive implements domain.UserRepository.
func (u *userRepository) ChangeActive(teamName string, userID string, isActive bool) (name string, err error) {
	reqCtx, cancel := context.WithTimeout(u.ctx, u.rtimeout)
	defer cancel()

	tx, err := u.pool.Begin(reqCtx)
	if err != nil {
		logrus.Error(logPrefix, err.Error())
		return "", err
	}
	defer tx.Rollback(reqCtx)

	var username string
	var user_active_state bool
	if err := tx.QueryRow(reqCtx, `SELECT name, is_active FROM users WHERE user_id=$1 AND team_name=$2`, userID, teamName).Scan(&username, &user_active_state); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", &errs.NotFoundError{
				Domain: "user",
			}
		}
		logrus.Error(logPrefix, err.Error())
		return "", &errs.InternalError{}
	}

	if user_active_state != isActive {
		if _, err := tx.Exec(reqCtx, `UPDATE users SET is_active = $1 WHERE team_name=$2 AND user_id=$3`, isActive, teamName, userID); err != nil {
			logrus.Error(logPrefix, "(update) error:", err.Error())
			return "", &errs.InternalError{}
		}
	}

	if err := tx.Commit(reqCtx); err != nil {
		logrus.Error(err.Error())
		return "", err
	}

	return username, nil
}
