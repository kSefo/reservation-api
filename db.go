package main

import (
	"database/sql"
	"time"

	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
)

// CreateDbMap DbMapを作成
func CreateDbMap(dbURL string) (*gorp.DbMap, error) {
	ds, err := createDatasource(dbURL)
	if err != nil {
		return nil, err
	}

	db := &gorp.DbMap{
		Db: ds,
		Dialect: gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "utf8mb4",
		},
	}

	db.AddTableWithName(Reservation{}, "reservation").SetKeys(true, "ID")
	return db, nil
}

func createDatasource(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(2)
	return db, nil
}

// Reservation 予約テーブルの構造体
type Reservation struct {
	ID                  uint      `db:"id"                   json:"id"`
	ReservationDatetime time.Time `db:"reservation_datetime" json:"reservation_datetime"`
	ReservationName     string    `db:"reservation_name"     json:"reservation_name"`
	Request             string    `db:"request"              json:"request"`
	Status              string    `db:"status"               json:"status"`
	Created             time.Time `db:"created"              json:"created"`
	Updated             time.Time `db:"updated"              json:"updated"`
}
