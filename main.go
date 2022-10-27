package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/nagohak/chat-app/config"
	"github.com/nagohak/chat-app/repository"
)

var addr = flag.String("addr", "localhost:8080", "http server address")
var redisAddr = flag.String("redisAddr", "localhost:6379", "redis url string")

func main() {
	flag.Parse()

	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Can't initialize database: %s", err)
	}
	defer db.Close()

	redis, err := config.NewRedis(*redisAddr)
	if err != nil {
		log.Fatalf("Can't initialize redis: %s", err)
	}

	fs := http.FileServer(http.Dir("./public"))

	userRepository := &repository.UserRepository{Db: db}
	roomRepository := &repository.RoomRepository{Db: db}

	ws := NewWsServer(roomRepository, userRepository, redis)
	go ws.Run()

	api := &Api{UserRepository: userRepository}

	http.Handle("/", fs)
	http.HandleFunc("/ws", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(ws, w, r)
	}))
	http.HandleFunc("/api/login", api.Login)

	log.Printf("Server is running on: %v", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
