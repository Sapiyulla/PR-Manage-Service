package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"pr-manage-service/internal/application/usecases"
	"pr-manage-service/internal/interfaces/handlers"
	"pr-manage-service/internal/transport/repository"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
)

type Mode int8

const (
	Debug Mode = 0
	Prod  Mode = 1
)

var (
	DSN  string
	MODE Mode = Debug //default

	ADMIN_TOKEN string
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "02.01.2006 15:04",
	})
	if mode := os.Getenv("MODE"); mode != "" {
		switch strings.ToLower(mode) {
		case "debug", "d":
			MODE = Debug
		case "prod", "production", "p":
			MODE = Prod
		}
	}
	if AdminToken := os.Getenv("ADMIN_TOKEN"); AdminToken != "" {
		ADMIN_TOKEN = AdminToken
		switch MODE {
		case Debug:
			log.Println("\033[33m(ENV)\033[39m ADMIN_TOKEN:\033[32m", AdminToken, "\033[39m")
		}
	}
	switch MODE {
	case Debug:
		gin.SetMode(gin.DebugMode)
	case Prod:
		gin.SetMode(gin.ReleaseMode)
	}
	if dsn := os.Getenv("DSN"); dsn == "" {
		log.Fatal("(ENV) DSN not setted")
	} else {
		if MODE == Debug {
			log.Println("\033[33m(ENV)\033[39m DSN:\033[32m", dsn, "\033[39m")
		}
		DSN = dsn
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poolCtx, poolCancel := context.WithTimeout(ctx, 3*time.Second)
	pool, err := pgxpool.New(poolCtx, DSN)
	if err != nil {
		log.Panicf("Error by pool init: %s", err.Error())
	}
	poolCancel()

	// team depends
	teamRepository := repository.NewTeamRepository(ctx, pool, 2*time.Second)
	teamUseCase := usecases.NewTeamUseCase(teamRepository)
	teamHandler := handlers.NewTeamHandler(teamUseCase)

	// pr depends
	prRepository := repository.NewPullRequestRepository(ctx, pool, 10*time.Second)
	prUseCase := usecases.NewPrUseCase(prRepository)
	prHandler := handlers.NewPRHandler(ctx, prUseCase)

	// user depends
	userRepository := repository.NewUserRepository(ctx, pool, 2*time.Second)
	userUseCase := usecases.NewUserUseCase(userRepository, prUseCase)
	userHandler := handlers.NewUserHandler(userUseCase, ADMIN_TOKEN)

	r := gin.Default()
	teamApi := r.Group("/team")
	{
		teamApi.POST("/add", teamHandler.AddTeamHandler)
		teamApi.GET("/get", teamHandler.GetTeamHandler)
	}
	userApi := r.Group("/users")
	{
		userApi.POST("/setIsActive", userHandler.SetIsActiveHandler)
		userApi.GET("/getReview", userHandler.GetReviewHandler)
	}
	prApi := r.Group("/pullRequest")
	{
		prApi.POST("create", prHandler.CreateHandler)
		prApi.POST("/merge", prHandler.MergeHandler)
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil {
			log.Fatal("(server) listening error: ", err.Error())
		}
	}()
	println("(server) listen http connections on ", server.Addr)

	wg.Wait()
	go func() {
		shdCtx, shdCencel := context.WithTimeout(ctx, 3*time.Second)
		defer shdCencel()
		if err := server.Shutdown(shdCtx); err != nil {
			log.Panicln("(server) shutdown error: ", err.Error())
		}
	}()
	time.Sleep(4 * time.Second)
	log.Println("(server) shutting downed")
}
