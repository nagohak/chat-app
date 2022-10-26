package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/nagohak/chat-app/auth"
	"github.com/nagohak/chat-app/config"
	"github.com/nagohak/chat-app/repository"
)

var addr = flag.String("addr", "localhost:8080", "http server address")

func main() {
	flag.Parse()

	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Can't initialize database: %s", err)
	}
	defer db.Close()

	config.NewRedis()

	fs := http.FileServer(http.Dir("./public"))

	userRepository := &repository.UserRepository{Db: db}
	roomRepository := &repository.RoomRepository{Db: db}

	ws := NewWsServer(roomRepository, userRepository)
	go ws.Run()

	api := &Api{UserRepository: userRepository}

	http.Handle("/", fs)
	http.HandleFunc("/ws", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(ws, w, r)
	}))
	http.HandleFunc("/api/login", api.Login)

	log.Printf("Server is running on: %v", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
