package service

import (
	"fmt"
	"net/mail"
	"net/smtp"
)

func (s *Service) sendMail(to, subject, body string) error {
	fromAddr := mail.Address{Address: s.noReply, Name: "norelpy"}
	toAddr := mail.Address{Address: to}
	var msg string
	msg += fmt.Sprintf("From: %s\r\n", fromAddr.String())
	msg += fmt.Sprintf("To: %s\r\n", toAddr.String())
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "Content-Type: text/html; charset=utf-8\r\n"
	msg += "\r\n"
	msg += body
	return smtp.SendMail(s.smtpAddr, s.smtpAuth, fromAddr.String(), []string{toAddr.String()}, []byte(msg))

}
