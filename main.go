package main

import (
	"log"
	"net/http"

	"github.com/nagohak/chat-app/api"
	"github.com/nagohak/chat-app/auth"
	"github.com/nagohak/chat-app/config"
	"github.com/nagohak/chat-app/database"
	"github.com/nagohak/chat-app/pkg/redis"
	"github.com/nagohak/chat-app/repository"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Can't initialize config: %s", err)
	}

	auth := auth.NewAuth()

	db, err := database.New(&database.Options{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		Db:       cfg.Postgres.Db,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
	}, auth)
	if err != nil {
		log.Fatalf("Can't initialize database: %s", err)
	}
	defer db.Close()

	redis, err := redis.New(cfg.Redis.Host, cfg.Redis.Port)
	if err != nil {
		log.Fatalf("Can't initialize redis: %s", err)
	}

	fs := http.FileServer(http.Dir("./public"))

	userRepository := repository.NewUserRepository(db)
	roomRepository := repository.NewRoomRepository(db)

	ws := NewWsServer(roomRepository, userRepository, redis)
	go ws.Run()

	api := api.NewApi(userRepository, auth)

	http.Handle("/", fs)
	http.HandleFunc("/ws", api.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(ws, w, r)
	}))
	http.HandleFunc("/api/login", api.Login)
	http.HandleFunc("/api/registration", api.Registration)

	log.Printf("Server is running on: %v", cfg.Http.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Http.Port, nil))
}
