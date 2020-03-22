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

// UserHandler Handlerの中でDBへクエリを発行するため、DB接続の構造体を持つ
type UserHandler struct {
	master *gorp.DbMap
}

// NewUserHandler Handlerの作成関数
func NewUserHandler(master *gorp.DbMap) http.Handler {
	return &UserHandler{
		master: master,
	}
}

// HTTPリクエストを受け、ビジネスロジックを実行してレスポンスを返す
func (h UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h UserHandler) serveGET(w http.ResponseWriter, r *http.Request) {
	result, err := h.master.Select(User{}, "SELECT * FROM user")
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Execute Query is failed").Write(w)
		return
	}

	userItems := make([]*User, 0)

	for _, e := range result {
		user := e.(*User)
		userItems = append(userItems, user)
	}

	data, err := json.Marshal(userItems)
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Marshal JSON is failed").Write(w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

// UserPostPayload （新規登録用）ユーザテーブルの構造体
type UserPostPayload struct {
	UserName  string `json:"user_name"`
	UserTel   string `json:"user_tel"`
	UserEmail string `json:"user_email"`
}

func (h UserHandler) servePOST(w http.ResponseWriter, r *http.Request) {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Read payload is failed").Write(w)
		return
	}

	var payload UserPostPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Parse payload is failed").Write(w)
		return
	}

	now := time.Now()
	user := User{
		UserName:  payload.UserName,
		UserTel:   payload.UserTel,
		UserEmail: payload.UserEmail,
		Created:   now,
		Updated:   now,
	}

	if err := h.master.Insert(&user); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Insert Data is failed").Write(w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
}

// UserPutPayload （更新用）ユーザテーブルの構造体
type UserPutPayload struct {
	UserID    uint   `json:"user_id"`
	UserName  string `json:"user_name"`
	UserTel   string `json:"user_tel"`
	UserEmail string `json:"user_email"`
}

func (h UserHandler) servePUT(w http.ResponseWriter, r *http.Request) {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Read payload is failed").Write(w)
		return
	}

	var payload UserPutPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Parse payload is failed").Write(w)
		return
	}

	var target User
	if err := h.master.SelectOne(&target, "SELECT * FROM user WHERE user_id = ?", payload.UserID); err != nil {
		if err == sql.ErrNoRows {
			NewErrorResponse(http.StatusNotFound, fmt.Sprintf("user_id=%d is not found", payload.UserID)).Write(w)
		} else {
			log.Println(err.Error())
			NewErrorResponse(http.StatusInternalServerError, "Select user is failed").Write(w)
		}
		return
	}

	now := time.Now()
	user := User{
		UserID:    payload.UserID,
		UserName:  payload.UserName,
		UserTel:   payload.UserTel,
		UserEmail: payload.UserEmail,
		Created:   target.Created,
		Updated:   now,
	}

	if _, err := h.master.Update(&user); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Update Data is failed").Write(w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
