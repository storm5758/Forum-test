package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	postgres "github.com/storm5758/Forum-test/internal/app/repository/postgres"
	"github.com/storm5758/Forum-test/internal/app/server"
	services "github.com/storm5758/Forum-test/internal/app/services"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// connection string
	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", Host, Port, User, Password, DBname)

	// connect to database
	pool, err := pgxpool.Connect(ctx, psqlConn)
	if err != nil {
		log.Fatal("can't connect to database", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("ping database error", err)
	}

	// настраиваем
	config := pool.Config()
	config.MaxConnIdleTime = MaxConnIdleTime
	config.MaxConnLifetime = MaxConnLifetime
	config.MinConns = MinConns
	config.MaxConns = MaxConns

	// ceate repository
	repo := postgres.NewRepository(pool)

	// create server
	srv, err := server.New(server.Services{
		Admin: services.NewAdminService(),
		User: services.NewUserService(services.Deps{
			UserRepository: repo,
		}),
		Forum:  services.NewForumService(),
		Post:   services.NewPostService(),
		Thread: services.NewThreadService(),
	})
	if err != nil {
		log.Fatalf("can't create server: %s", err.Error())
	}

	// run server
	if err := srv.Run(ctx); err != nil {
		log.Println(err)
	}
}
