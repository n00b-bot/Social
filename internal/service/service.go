package service

import (
	"database/sql"
	"net"
	"net/smtp"
	"sync"
)

type Service struct {
	db                  *sql.DB
	tokenKey            string
	noReply             string
	smtpAddr            string
	smtpAuth            smtp.Auth
	timelineItemClients sync.Map
	commentClient       sync.Map
	notificationClient  sync.Map
}

type Config struct {
	Db       *sql.DB
	Secret   string
	SMTPHost string
	SMTPPort string
	SMTPuser string
	SMTPPass string
}

func New(cfg Config) *Service {
	service := &Service{
		db:       cfg.Db,
		tokenKey: cfg.Secret,
		smtpAddr: net.JoinHostPort(cfg.SMTPHost, cfg.SMTPPort),
		smtpAuth: smtp.PlainAuth("", cfg.SMTPuser, cfg.SMTPPass, cfg.SMTPHost),
		noReply:  "noreply@" + "dot.com",
	}

	return service
}
