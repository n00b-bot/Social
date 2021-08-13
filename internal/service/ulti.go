package service

import "github.com/go-sql-driver/mysql"

func isUniqueViolation(err error) bool {
	mysqll, ok := err.(*mysql.MySQLError)
	return ok && mysqll.Message == "..."
}
