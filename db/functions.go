package db

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

type validGenerics interface {
}

func Query[T validGenerics](db *DB, query string, args map[string]any) ([]T, error) {
	rows, err := db.Conn.NamedQuery(query, args)
	if err != nil {
		err = fmt.Errorf("error exequting query >> %w", err)
		return nil, err
	}
	defer rows.Close()

	results := []T{}
	for rows.Next() {
		var item T
		err := rows.StructScan(&item)
		if err != nil {
			err = fmt.Errorf("error scanning row >> %w", err)
			return nil, err
		}
		results = append(results, item)
	}

	return results, nil
}

func Exec(db *DB, query string, args map[string]any) (result string, err error) {
	hasReturning := strings.Contains(strings.ToLower(query), "returning")

	if hasReturning {
		var id string

		rows, errQuery := db.Conn.NamedQuery(query, args)
		if errQuery != nil {
			err = fmt.Errorf("error executing query >> %w", errQuery)
			return
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&id)
			if err != nil {
				err = fmt.Errorf("error scanning returned id >> %w", err)
				return
			}
			result = id
		}
	} else {
		namedQuery, namedArgs, errNamed := sqlx.Named(query, args)
		if errNamed != nil {
			err = fmt.Errorf("error preparing named query >> %w", errNamed)
			return
		}

		namedQuery = db.Conn.Rebind(namedQuery)

		_, errExec := db.Conn.Exec(namedQuery, namedArgs...)
		if errExec != nil {
			err = fmt.Errorf("error executing query >> %w", errExec)
			return
		}
	}

	return
}

func In(query string, args map[string]any) (newQuery string, newArgs map[string]any) {
	for index, value := range args {
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice {
			continue
		}

		var queryValues []string
		for i := 0; i < rv.Len(); i++ {
			key := fmt.Sprintf("%s%d", index, i)
			args[key] = rv.Index(i).Interface()
			queryValues = append(queryValues, fmt.Sprintf(":%s", key))
		}

		query = strings.ReplaceAll(query, ":"+index, strings.Join(queryValues, ", "))

		delete(args, index)
	}

	newQuery = query
	newArgs = args

	return
}
