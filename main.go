package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"network/internal/handler"
	"network/internal/service"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hako/branca"
)

func main() {

	db, err := sql.Open("mysql", "root:root@/backend")
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
	s := service.New(db, codec)
	h := handler.New(s)
	http.ListenAndServe(":3000", h)
}
