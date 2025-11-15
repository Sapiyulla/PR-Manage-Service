package repository

import (
	"context"
	"errors"
	"pr-manage-service/internal/domain"
	"pr-manage-service/pkg/errs"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

const logPrefix = "(database) error:"

type teamRepository struct {
	ctx     context.Context
	pool    *pgxpool.Pool
	timeout time.Duration
}

func NewTeamRepository(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration) domain.TeamRepository {
	return &teamRepository{
		ctx:     ctx,
		pool:    pool,
		timeout: timeout,
	}
}

// AddNewTeam implements domain.TeamRepository.
func (t *teamRepository) AddNewTeam(teamName string, members *[]domain.User) error {
	reqCtx, cancel := context.WithTimeout(t.ctx, t.timeout)
	defer cancel()

	tx, err := t.pool.Begin(reqCtx)
	if err != nil {
		logrus.Error(logPrefix, err.Error())
		return &errs.InternalError{}
	}
	defer tx.Rollback(reqCtx)

	// 1. Вставляем команду
	_, err = tx.Exec(reqCtx, `INSERT INTO teams (name) VALUES ($1)`, teamName)
	if err != nil {
		if strings.Contains(err.Error(), "dublicate") || strings.Contains(err.Error(), "duplicate") {
			return &errs.AlreadyExistsError{
				Domain: "team_name",
			}
		}
		logrus.Error(logPrefix, "(insert team) error: ", err.Error())
		return &errs.InternalError{}
	}

	// 2. Вставляем пользователей с использованием плейсхолдеров
	if len(*members) > 0 {
		// Подготовка пакетного запроса
		batch := &pgx.Batch{}

		for _, member := range *members {
			batch.Queue(
				`INSERT INTO users (user_id, team_name, name, is_active) VALUES ($1, $2, $3, $4)`,
				member.UserID, teamName, member.UserName, member.IsActive,
			)
		}

		// Выполняем пакет
		br := tx.SendBatch(reqCtx, batch)
		defer br.Close()

		// Проверяем результаты каждого запроса в пакете
		for i := 0; i < batch.Len(); i++ {
			_, err := br.Exec()
			if err != nil {
				logrus.Error(logPrefix, "(batch exec) error at query %d: %v\n", i, err)
				return &errs.InternalError{}
			}
		}

		// Закрываем пакет
		if err := br.Close(); err != nil {
			logrus.Error(logPrefix, "(batch close) error: ", err.Error())
			return &errs.InternalError{}
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(reqCtx); err != nil {
		logrus.Error(logPrefix, "(tx.commit) error: ", err.Error())
		return &errs.InternalError{}
	}

	return nil
}

func ParallelExecute(ctx context.Context, wg *sync.WaitGroup, tx pgx.Tx, sql string, args ...any) {
	defer wg.Done()
	if _, err := tx.Exec(ctx, sql, args...); err != nil {
		logrus.Error(logPrefix, "(tx) error: ", err.Error())
		if err := tx.Rollback(ctx); err != nil {
			logrus.Error(logPrefix, "(tx.rollback) error: ", err.Error())
			return
		}
	}
}

// GetTeamInfoByName implements domain.TeamRepository.
func (t *teamRepository) GetTeamInfoByName(teamName string) (*domain.Team, error) {
	reqCtx, cancel := context.WithTimeout(t.ctx, t.timeout)
	defer cancel()
	tx, err := t.pool.Begin(reqCtx)
	if err != nil {
		return nil, err
	}
	var team domain.Team = domain.Team{Members: make([]domain.User, 0, 10)}
	row := tx.QueryRow(reqCtx, `SELECT name FROM teams WHERE name=$1`, teamName)
	if err := row.Scan(&team.TeamName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &errs.NotFoundError{
				Domain: "team",
			}
		}
		logrus.Error(logPrefix, err.Error())
		return nil, &errs.InternalError{}
	}

	rows, err := tx.Query(reqCtx,
		`SELECT user_id, name, is_active FROM users 
		WHERE team_name = $1`, teamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &team, nil
		}
		logrus.Error(logPrefix, err.Error())
		return nil, &errs.InternalError{}
	}
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.UserName, &user.IsActive); err != nil {
			logrus.Error(logPrefix, "(row scan) error:", err.Error())
			continue
		}
		team.Members = append(team.Members, user)
	}

	return &team, nil
}
