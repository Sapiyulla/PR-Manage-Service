package main

import (
	"context"
	"log"
	"net/http"
	"os"
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
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "02.01.2006 15:04",
	})
	if mode := os.Getenv("MODE"); mode != "" {
		switch strings.ToLower(mode) {
		case "debug":
			MODE = Debug
		case "prod", "production", "p":
			MODE = Prod
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
			log.Println("\033[33(mENV) \033[32mDSN:", dsn, "\033[39m")
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
	_ = pool
	//teamHandler := handlers.NewTeamHandler()

	r := gin.Default()
	// teamApi := r.Group("/team")
	// {
	// 	teamApi.POST("/add", teamHandler.AddTeamHandler)
	// 	teamApi.GET("/get", teamHandler.GetTeamHandler)
	// }

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
