package main

import (
	"database/sql"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
)

type DB interface {
	Exec(sql string, vars ...interface{}) (sql.Result, error)
	Query(sql string, vars ...interface{}) (Rows, error)
	QueryRow(sql string, vars ...interface{}) Row
	Begin() (Tx, error)
}

type Tx interface {
	Exec(sql string, vars ...interface{}) (sql.Result, error)
	Query(sql string, vars ...interface{}) (Rows, error)
	QueryRow(sql string, vars ...interface{}) Row
	Rollback() error
	Commit() error
}

type Row interface {
	Scan(out ...interface{}) error
}

type Rows interface {
	Next() bool
	Scan(out ...interface{}) error
	Close() error
}

type queryWatcher func(begin time.Time, dur time.Duration, name, file string, line int, transactionID string, sql string, vars []interface{})

type dbLogger struct {
	db *sql.DB
	l  []queryWatcher
	n  snowflake.Node
}

func (d *dbLogger) location(skip int) (string, string, int) {
	if pc, file, line, ok := runtime.Caller(skip); ok {
		p := strings.Index(file, "fknsrs.biz/p/don")
		return runtime.FuncForPC(pc).Name(), file[p:], line
	}

	return "fake_function", "fake_file", -1
}

func (d *dbLogger) Query(sql string, vars ...interface{}) (Rows, error) {
	before := time.Now()

	name, file, line := d.location(2)

	var rl *rowsLogger
	rows, err := d.db.Query(sql, vars...)
	if rows != nil {
		rl = &rowsLogger{Rows: rows}

		rl.AfterClose(func(_ *rowsLogger) {
			dur := time.Since(before)

			for _, fn := range d.l {
				fn(before, dur, name, file, line, "", sql, vars)
			}
		})
	}

	return rl, err
}

func (d *dbLogger) QueryRow(sql string, vars ...interface{}) Row {
	before := time.Now()

	name, file, line := d.location(2)

	rl := &rowLogger{Row: d.db.QueryRow(sql, vars...)}

	rl.AfterScan(func(_ *rowLogger) {
		dur := time.Since(before)

		for _, fn := range d.l {
			fn(before, dur, name, file, line, "", sql, vars)
		}
	})

	return rl
}

func (d *dbLogger) Exec(sql string, vars ...interface{}) (sql.Result, error) {
	before := time.Now()

	name, file, line := d.location(2)

	defer func() {
		dur := time.Since(before)

		for _, fn := range d.l {
			fn(before, dur, name, file, line, "", sql, vars)
		}
	}()

	return d.db.Exec(sql, vars...)
}

func (d *dbLogger) Begin() (Tx, error) {
	id := d.n.Generate()

	before := time.Now()

	name, file, line := d.location(2)

	defer func() {
		dur := time.Since(before)

		for _, fn := range d.l {
			fn(before, dur, name, file, line, id.String(), "BEGIN", nil)
		}
	}()

	tx, err := d.db.Begin()
	return &txLogger{id: id, db: d, tx: tx}, err
}

type txLogger struct {
	id   snowflake.ID
	db   *dbLogger
	tx   *sql.Tx
	done bool
	err  error
}

func (t *txLogger) Commit() error {
	if t.done {
		return t.err
	}

	before := time.Now()

	name, file, line := t.db.location(2)

	defer func() {
		dur := time.Since(before)

		for _, fn := range t.db.l {
			fn(before, dur, name, file, line, t.id.String(), "COMMIT", nil)
		}
	}()

	t.done = true
	t.err = t.tx.Commit()

	return t.err
}

func (t *txLogger) Rollback() error {
	if t.done {
		return t.err
	}

	before := time.Now()

	name, file, line := t.db.location(2)

	defer func() {
		dur := time.Since(before)

		for _, fn := range t.db.l {
			fn(before, dur, name, file, line, t.id.String(), "ROLLBACK", nil)
		}
	}()

	t.done = true
	t.err = t.tx.Rollback()

	return t.err
}

func (t *txLogger) Query(sql string, vars ...interface{}) (Rows, error) {
	before := time.Now()

	name, file, line := t.db.location(2)

	var rl *rowsLogger
	rows, err := t.tx.Query(sql, vars...)
	if rows != nil {
		rl = &rowsLogger{Rows: rows}

		rl.AfterClose(func(_ *rowsLogger) {
			dur := time.Since(before)

			for _, fn := range t.db.l {
				fn(before, dur, name, file, line, t.id.String(), sql, vars)
			}
		})
	}

	return rl, err
}

func (t *txLogger) QueryRow(sql string, vars ...interface{}) Row {
	before := time.Now()

	name, file, line := t.db.location(2)

	rl := &rowLogger{Row: t.tx.QueryRow(sql, vars...)}

	rl.AfterScan(func(_ *rowLogger) {
		dur := time.Since(before)

		for _, fn := range t.db.l {
			fn(before, dur, name, file, line, t.id.String(), sql, vars)
		}
	})

	return rl
}

func (t *txLogger) Exec(sql string, vars ...interface{}) (sql.Result, error) {
	before := time.Now()

	name, file, line := t.db.location(2)

	defer func() {
		dur := time.Since(before)

		for _, fn := range t.db.l {
			fn(before, dur, name, file, line, t.id.String(), sql, vars)
		}
	}()

	return t.tx.Exec(sql, vars...)
}

type rowLogger struct {
	*sql.Row
	fns  []func(r *rowLogger)
	done bool
	err  error
}

func (r *rowLogger) AfterScan(fn func(r *rowLogger)) {
	r.fns = append(r.fns, fn)
}

func (r *rowLogger) Scan(vars ...interface{}) error {
	if !r.done {
		r.done = true
		r.err = r.Row.Scan(vars...)

		for _, fn := range r.fns {
			fn(r)
		}
	}

	return r.err
}

type rowsLogger struct {
	*sql.Rows
	fns  []func(r *rowsLogger)
	done bool
	err  error
}

func (r *rowsLogger) AfterClose(fn func(r *rowsLogger)) {
	r.fns = append(r.fns, fn)
}

func (r *rowsLogger) Close() error {
	if !r.done {
		r.done = true
		r.err = r.Rows.Close()

		for _, fn := range r.fns {
			fn(r)
		}
	}

	return r.err
}

func printQuery(sqlString string, vars []interface{}) string {
	re := regexp.MustCompile(`\$([0-9]+)`)
	ws := regexp.MustCompile(`\s+`)

	return strings.TrimSpace(ws.ReplaceAllString(re.ReplaceAllStringFunc(sqlString, func(s string) string {
		i, err := strconv.ParseInt(s[1:], 10, 64)
		if err != nil {
			return s
		}

		if i < 1 || int(i) > len(vars) {
			return s
		}

		switch e := vars[i-1].(type) {
		case sql.NullBool:
			if !e.Valid {
				return "NULL"
			}

			return fmt.Sprintf("%v", e.Bool)
		case *sql.NullBool:
			if !e.Valid {
				return "NULL"
			}

			return fmt.Sprintf("%v", e.Bool)
		case sql.NullFloat64:
			if !e.Valid {
				return "NULL"
			}

			return fmt.Sprintf("%v", e.Float64)
		case *sql.NullFloat64:
			if !e.Valid {
				return "NULL"
			}

			return fmt.Sprintf("%v", e.Float64)
		case sql.NullInt64:
			if !e.Valid {
				return "NULL"
			}

			return fmt.Sprintf("%v", e.Int64)
		case *sql.NullInt64:
			if !e.Valid {
				return "NULL"
			}

			return fmt.Sprintf("%v", e.Int64)
		case sql.NullString:
			if !e.Valid {
				return "NULL"
			}

			return fmt.Sprintf("'%v'", e.String)
		case *sql.NullString:
			if !e.Valid {
				return "NULL"
			}

			return fmt.Sprintf("'%v'", e.String)
		default:
			return fmt.Sprintf("'%v'", vars[i-1])
		}
	}), " "))
}
