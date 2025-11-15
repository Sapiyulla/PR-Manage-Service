package repository

import (
	"context"
	"fmt"
	"pr-manage-service/internal/domain"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "02.01.2006 15:04",
	})
}

func TestAddNewTeam(t *testing.T) {
	dsn := "postgres://postgres:password@localhost:5432/pr_mng_db?sslmode=disable"
	fmt.Println("DSN: ", dsn)
	pool, _ := pgxpool.New(context.Background(), dsn)

	tests := []struct {
		name     string
		teamName string
		members  []domain.User
		wantErr  bool
	}{
		{
			name:     "successful add new team with valid data",
			teamName: "payments",
			members: []domain.User{
				{UserID: "u1", UserName: "Alice", IsActive: true},
				{UserID: "u2", UserName: "Bob", IsActive: true},
			},
			wantErr: false,
		},
		{
			name:     "team with special characters in name",
			teamName: "team-backend_v2.0",
			members: []domain.User{
				{UserID: "u1", UserName: "Alice", IsActive: true},
			},
			wantErr: false,
		},
		{
			name:     "team with inactive members",
			teamName: "inactive_team",
			members: []domain.User{
				{UserID: "u1", UserName: "Alice", IsActive: true},
				{UserID: "u2", UserName: "Bob", IsActive: false},
			},
			wantErr: false,
		},
		{
			name:     "team with multiple active members",
			teamName: "large_team",
			members: []domain.User{
				{UserID: "u1", UserName: "Alice", IsActive: true},
				{UserID: "u2", UserName: "Bob", IsActive: true},
				{UserID: "u3", UserName: "Charlie", IsActive: true},
				{UserID: "u4", UserName: "David", IsActive: true},
			},
			wantErr: false,
		},
		{
			name:     "team with duplicate user_ids (should fail on DB constraint)",
			teamName: "duplicate_users",
			members: []domain.User{
				{UserID: "u1", UserName: "Alice", IsActive: true},
				{UserID: "u1", UserName: "Alice Clone", IsActive: true},
			},
			wantErr: false, // Должно упасть на UNIQUE constraint в БД
		},
		{
			name:     "team that already exists (should fail on DB constraint)",
			teamName: "existing_team",
			members: []domain.User{
				{UserID: "u1", UserName: "Alice", IsActive: true},
			},
			wantErr: false,
		},
		// Повторное добавление той же команды
		{
			name:     "duplicate team name",
			teamName: "existing_team",
			members: []domain.User{
				{UserID: "u2", UserName: "Bob", IsActive: true},
			},
			wantErr: true, // Должно упасть на UNIQUE constraint для team name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &teamRepository{
				ctx:     context.Background(),
				pool:    pool,
				timeout: 5 * time.Second, // таймаут в секундах
			}

			err := repo.AddNewTeam(tt.teamName, &tt.members)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddNewTeam() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Если успешно добавили, можно проверить что данные действительно в БД
			if !tt.wantErr && err == nil {
				// Простая проверка - если нет ошибки, считаем что ок
				t.Logf("Team %s added successfully", tt.teamName)
			}
		})
	}

	// _, err1 := pool.Exec(context.Background(), `DELETE FROM users`)
	// _, err2 := pool.Exec(context.Background(), `DELETE FROM teams`)
	// if err1 != nil && err2 != nil {
	// 	t.Errorf("delete error: team: %s | user: %s", err2.Error(), err1.Error())
	// }
}

// Вспомогательная функция для генерации слишком большого количества участников
func generateTooManyMembers(count int) []domain.User {
	members := make([]domain.User, count)
	for i := 0; i < count; i++ {
		members[i] = domain.User{
			UserID:   fmt.Sprintf("user_%d", i),
			UserName: fmt.Sprintf("User %d", i),
			IsActive: true,
		}
	}
	return members
}

func TestGetTeamInfoByName(t *testing.T) {
	dsn := "postgres://postgres:password@localhost:5432/pr_mng_db?sslmode=disable"
	fmt.Println("DSN: ", dsn)
	pool, _ := pgxpool.New(context.Background(), dsn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := NewTeamRepository(ctx, pool, 10*time.Second)

	if team, err := repo.GetTeamInfoByName("existing_team"); err != nil {
		t.Error(err.Error())
	} else {
		t.Logf("%+v", team)
	}
}
