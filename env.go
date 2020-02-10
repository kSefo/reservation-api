package main

import (
	"errors"
	"os"
)

// Env 環境変数の構造体
type Env struct {
	Bind      string
	MasterURL string
}

// CreateEnv 環境変数を取得し構造体を作成
func CreateEnv() (*Env, error) {
	env := Env{}

	// APIをListenするポート設定
	bind := os.Getenv("RESERVATION_BIND")
	if bind == "" {
		env.Bind = ":8080"
	}
	env.Bind = bind

	// MySQL Masterへの接続情報
	masterURL := os.Getenv("RESERVATION_MASTER_URL")
	if masterURL == "" {
		return nil, errors.New("RESERVATION_MASTER_URL is not specified")
	}
	env.MasterURL = masterURL

	return &env, nil
}
