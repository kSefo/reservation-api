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

	db.AddTableWithName(SalesDay{}, "sales_day").SetKeys(true, "SalesDay")
	db.AddTableWithName(User{}, "user").SetKeys(true, "UserID")
	db.AddTableWithName(Menu{}, "menu").SetKeys(true, "MenuID")
	db.AddTableWithName(Reservation{}, "reservation").SetKeys(true, "ReservationID")
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

// SalesDay 営業日テーブルの構造体
type SalesDay struct {
	SalesDay  string    `db:"sales_day"  json:"sales_day"`
	StartTime string    `db:"start_time" json:"start_time"`
	EndTime   string    `db:"end_time"   json:"end_time"`
	Holiday   bool      `db:"holiday"    json:"holiday"`
	Created   time.Time `db:"created"    json:"created"`
	Updated   time.Time `db:"updated"    json:"updated"`
}

// User 利用者テーブルの構造体
type User struct {
	UserID    uint      `db:"user_id"    json:"user_id"`
	UserName  string    `db:"user_name"  json:"user_name"`
	UserTel   string    `db:"user_tel"   json:"user_tel"`
	UserEmail string    `db:"user_email" json:"user_email"`
	Created   time.Time `db:"created"    json:"created"`
	Updated   time.Time `db:"updated"    json:"updated"`
}

// Menu メニューテーブルの構造体
type Menu struct {
	MenuID   uint      `db:"menu_id"    json:"menu_id"`
	MenuName string    `db:"menu_name"  json:"menu_name"`
	Created  time.Time `db:"created"    json:"created"`
	Updated  time.Time `db:"updated"    json:"updated"`
}

// Reservation 予約テーブルの構造体
type Reservation struct {
	ReservationID       uint      `db:"reservation_id"       json:"reservation_id"`
	UserID              uint      `db:"user_id"              json:"user_id"`
	MenuID              uint      `db:"menu_id"              json:"menu_id"`
	ReservationDateTime time.Time `db:"reservation_datetime" json:"reservation_datetime"`
	Created             time.Time `db:"created"              json:"created"`
	Updated             time.Time `db:"updated"              json:"updated"`
}
