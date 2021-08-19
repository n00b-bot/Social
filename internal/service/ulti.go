package service

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/go-sql-driver/mysql"
)

const (
	miniPageSize    = 1
	defaultPageSize = 10
	maxPageSize     = 99
)

var queriesCache = make(map[string]*template.Template)
var rxMentions = regexp.MustCompile(`\B@([a-zA-Z][a-zA-Z0-9_-]{0,17})`)

func isUniqueViolation(err error) bool {
	mysqll, ok := err.(*mysql.MySQLError)
	return ok && mysqll.Number == 1062
}

func buildQuery(text string, data map[string]interface{}) (string, []interface{}, error) {
	var t *template.Template
	t, ok := queriesCache[text]
	if !ok {
		var err error
		t, err = template.New("query").Parse(text)
		if err != nil {
			return "", nil, err
		}
		queriesCache[text] = t

	}
	var wr bytes.Buffer
	if err := t.Execute(&wr, data); err != nil {
		return "", nil, fmt.Errorf("could not apply sql query data: %w", err)
	}
	query := wr.String()
	var args []interface{}
	for key, val := range data {
		if !strings.Contains(query, "@"+key) {
			continue
		}
		args = append(args, val)
		query = strings.ReplaceAll(query, "@"+key, fmt.Sprintf("$%d", len(args)))
	}

	return query, args, nil
}

func normalizePageSize(i int) int {
	if i == 0 {
		return defaultPageSize
	}
	if i < miniPageSize {
		return miniPageSize
	}
	if i > maxPageSize {
		return maxPageSize
	}
	return i
}

func collectionMentions(s string) []string {
	m := map[string]struct{}{}
	u := []string{}
	for _, submatches := range rxMentions.FindAllStringSubmatch(s, -1) {
		val := submatches[1]
		if _, ok := m[val]; !ok {
			m[val] = struct{}{}
			u = append(u, val)
		}
	}
	return u
}
