package service

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
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
	var ints []int
	for key, _ := range data {
		if !strings.Contains(query, "@"+key) {
			continue
		}
		i, _ := strconv.Atoi(key[1:])
		ints = append(ints, i)
		query = strings.ReplaceAll(query, "@"+key, "?")
	}
	sort.Ints(ints)
	var args []interface{}
	for _, key := range ints {

		args = append(args, data["a"+strconv.Itoa(key)])
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
