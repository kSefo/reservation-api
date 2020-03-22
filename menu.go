package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-gorp/gorp"
)

// MenuHandler Handlerの中でDBへクエリを発行するため、DB接続の構造体を持つ
type MenuHandler struct {
	master *gorp.DbMap
}

// NewMenuHandler Handlerの作成関数
func NewMenuHandler(master *gorp.DbMap) http.Handler {
	return &MenuHandler{
		master: master,
	}
}

// HTTPリクエストを受け、ビジネスロジックを実行してレスポンスを返す
func (h MenuHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] RemoteAddr=%s\tUserAgent=%s", r.Method, r.RemoteAddr, r.Header.Get("User-Agent"))
	switch r.Method {
	case "GET":
		h.serveGET(w, r)
		return
	case "POST":
		h.servePOST(w, r)
		return
	case "PUT":
		h.servePUT(w, r)
		return
	default:
		NewErrorResponse(http.StatusMethodNotAllowed, fmt.Sprintf("%s is Unsupported method", r.Method)).Write(w)
		return
	}
}

func (h MenuHandler) serveGET(w http.ResponseWriter, r *http.Request) {
	result, err := h.master.Select(Menu{}, "SELECT * FROM menu")
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Execute Query is failed").Write(w)
		return
	}

	menuItems := make([]*Menu, 0)

	for _, e := range result {
		menu := e.(*Menu)
		menuItems = append(menuItems, menu)
	}

	data, err := json.Marshal(menuItems)
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Marshal JSON is failed").Write(w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

// MenuPostPayload （新規登録用）メニューテーブルの構造体
type MenuPostPayload struct {
	MenuName string `json:"menu_name"`
}

func (h MenuHandler) servePOST(w http.ResponseWriter, r *http.Request) {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Read payload is failed").Write(w)
		return
	}

	var payload MenuPostPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Parse payload is failed").Write(w)
		return
	}

	now := time.Now()
	menu := Menu{
		MenuName: payload.MenuName,
		Created:  now,
		Updated:  now,
	}

	if err := h.master.Insert(&menu); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Insert Data is failed").Write(w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
}

// MenuPutPayload （更新用）メニューテーブルの構造体
type MenuPutPayload struct {
	MenuID   uint   `json:"menu_id"`
	MenuName string `json:"menu_name"`
}

func (h MenuHandler) servePUT(w http.ResponseWriter, r *http.Request) {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Read payload is failed").Write(w)
		return
	}

	var payload MenuPutPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Parse payload is failed").Write(w)
		return
	}

	var target Menu
	if err := h.master.SelectOne(&target, "SELECT * FROM menu WHERE menu_id = ?", payload.MenuID); err != nil {
		if err == sql.ErrNoRows {
			NewErrorResponse(http.StatusNotFound, fmt.Sprintf("menu_id=%d is not found", payload.MenuID)).Write(w)
		} else {
			log.Println(err.Error())
			NewErrorResponse(http.StatusInternalServerError, "Select menu is failed").Write(w)
		}
		return
	}

	now := time.Now()
	menu := Menu{
		MenuID:   payload.MenuID,
		MenuName: payload.MenuName,
		Created:  target.Created,
		Updated:  now,
	}

	if _, err := h.master.Update(&menu); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Update Data is failed").Write(w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
