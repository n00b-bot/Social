package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"network/internal/handler"
	"network/internal/service"

	"github.com/hako/branca"
	_ "github.com/lib/pq"
)

func main() {

	db, err := sql.Open("postgres", "postgresql://root@127.0.0.1:26257/nakama?sslmode=disable")
	if err != nil {
		fmt.Print(err)
		return
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		fmt.Print(err)
		return
	}
	codec := branca.NewBranca("12345678123456781234567812345678")
	codec.SetTTL(uint32(service.TokenLifespan.Seconds()))
	s := service.New(db, codec)
	h := handler.New(s)
	http.ListenAndServe(":3000", h)
}
