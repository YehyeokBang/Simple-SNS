package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/YehyeokBang/Simple-SNS/config"
	"github.com/YehyeokBang/Simple-SNS/pkg/db"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := db.NewGormDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	fmt.Println(db)

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	http.ListenAndServe(":8080", nil)
}
