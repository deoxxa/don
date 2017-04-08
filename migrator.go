package main

import (
	"database/sql"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/Sirupsen/logrus"
)

func migrate(db *sql.DB, box *rice.Box) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`create table if not exists migrations (id numeric key, name text not null unique, applied_at datetime not null);`); err != nil {
		return err
	}

	var names []string
	if err := box.Walk("", func(p string, m os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(p, ".sql") {
			names = append(names, p)
		}

		return nil
	}); err != nil {
		return err
	}
	sort.Strings(names)

	for _, n := range names {
		s, err := box.String(n)
		if err != nil {
			return err
		}

		logrus.WithField("file", n).Info("checking migration status")

		var count int
		if err := tx.QueryRow("select count(1) from migrations where name = $1", n).Scan(&count); err != nil {
			return err
		} else if count == 0 {
			logrus.WithField("file", n).Info("applying migration")

			if _, err := tx.Exec(s); err != nil {
				return err
			}

			if _, err = tx.Exec("insert into migrations (name, applied_at) values($1, $2)", n, time.Now()); err != nil {
				return err
			}
		} else {
			logrus.WithField("file", n).Info("already applied")
		}
	}

	return tx.Commit()
}
