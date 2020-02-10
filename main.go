package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// 環境変数を格納した構造体を作成
	env, err := CreateEnv()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}

	// MySQL Masterへの接続するための構造体を作成
	masterDB, err := CreateDbMap(env.MasterURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s is invalid database", env.MasterURL)
		return
	}

	mux := http.NewServeMux()

	// ヘルスチェック用APIのハンドラを作成
	hc := func(w http.ResponseWriter, r *http.Request) {
		log.Println("[GET] /hc")
		w.Write([]byte("OK"))
	}

	// 予約テーブル操作APIのハンドラを作成
	reservationHandler := NewReservationHandler(masterDB)

	// ハンドラをAPIエンドポイントとして登録
	mux.Handle("/reservation", reservationHandler)
	mux.HandleFunc("/hc", hc)

	// サーバのポートやハンドラを設定し、Listenを開始
	s := http.Server{
		Addr:    env.Bind,
		Handler: mux,
	}
	log.Printf("Listen HTTP Server")
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
