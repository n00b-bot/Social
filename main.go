package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"network/internal/handler"
	"network/internal/service"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	db, err := sql.Open("postgres", "postgresql://root@127.0.0.1:26257/nakama?sslmode=disable")
	if err != nil {
		fmt.Print(err)
		return
	}

	var (
		SMTP_HOST = env("SMTP_HOST", "smtp.mailtrap.io")
		SMTP_PORT = env("SMTP_PORT", "465")
		SMTP_User = env("SMTP_USERNAME", "")
		SMTP_Pass = env("SMTP_PASSWORD", "")
	)
	defer db.Close()
	if err = db.Ping(); err != nil {
		fmt.Print(err)
		return
	}

	s := service.New(service.Config{
		Db:       db,
		Secret:   "12345678123456781234567812345678",
		SMTPHost: SMTP_HOST,
		SMTPPort: SMTP_PORT,
		SMTPuser: SMTP_User,
		SMTPPass: SMTP_Pass,
	})
	h := handler.New(s)
	http.ListenAndServe(":3000", h)
}

func env(key, defaultValue string) string {
	s, ok := os.LookupEnv(key)
	if ok {
		return s
	}
	return defaultValue
}
