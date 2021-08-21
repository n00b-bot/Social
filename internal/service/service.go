package service

import (
	"database/sql"
	"net"
	"net/smtp"
	"sync"

	"github.com/hako/branca"
)

type Service struct {
	db                  *sql.DB
	codec               *branca.Branca
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
	codec := branca.NewBranca(cfg.Secret)
	codec.SetTTL(uint32(TokenLifespan.Seconds()))
	service := &Service{
		db:       cfg.Db,
		codec:    codec,
		smtpAddr: net.JoinHostPort(cfg.SMTPHost, cfg.SMTPPort),
		smtpAuth: smtp.PlainAuth("", cfg.SMTPuser, cfg.SMTPPass, cfg.SMTPHost),
		noReply:  "noreply@" + "dot.com",
	}

	return service
}
