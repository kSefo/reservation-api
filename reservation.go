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

// ReservationHandler Handlerの中でDBへクエリを発行するため、DB接続の構造体を持つ
type ReservationHandler struct {
	master *gorp.DbMap
}

// NewReservationHandler Handlerの作成関数
func NewReservationHandler(master *gorp.DbMap) http.Handler {
	return &ReservationHandler{
		master: master,
	}
}

// HTTPリクエストを受け、ビジネスロジックを実行してレスポンスを返す
func (h ReservationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h ReservationHandler) serveGET(w http.ResponseWriter, r *http.Request) {
	reservationDateFrom := r.URL.Query().Get("reservationDateFrom")
	reservationDateTo := r.URL.Query().Get("reservationDateTo")
	result, err := h.master.Select(Reservation{}, "SELECT * FROM reservation WHERE reservation_datetime BETWEEN ? AND ? ORDER BY reservation_datetime", reservationDateFrom+" 00:00:00", reservationDateTo+" 23:59:59")
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Execute Query is failed").Write(w)
		return
	}

	reservationItems := make([]*Reservation, 0)

	for _, e := range result {
		reservation := e.(*Reservation)
		reservationItems = append(reservationItems, reservation)
	}

	data, err := json.Marshal(reservationItems)
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Marshal JSON is failed").Write(w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

// ReservationPostPayload （新規登録用）予約テーブルの構造体
type ReservationPostPayload struct {
	UserID              uint      `json:"user_id"`
	MenuID              uint      `json:"menu_id"`
	ReservationDateTime time.Time `json:"reservation_datetime"`
}

func (h ReservationHandler) servePOST(w http.ResponseWriter, r *http.Request) {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Read payload is failed").Write(w)
		return
	}

	var payload ReservationPostPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Parse payload is failed").Write(w)
		return
	}

	now := time.Now()
	reservation := Reservation{
		UserID:              payload.UserID,
		MenuID:              payload.MenuID,
		ReservationDateTime: payload.ReservationDateTime,
		Created:             now,
		Updated:             now,
	}

	if err := h.master.Insert(&reservation); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Insert Data is failed").Write(w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
}

// ReservationPutPayload （更新用）予約テーブルの構造体
type ReservationPutPayload struct {
	ReservationID       uint      `json:"reservation_id"`
	UserID              uint      `json:"user_id"`
	MenuID              uint      `json:"menu_id"`
	ReservationDateTime time.Time `json:"reservation_datetime"`
}

func (h ReservationHandler) servePUT(w http.ResponseWriter, r *http.Request) {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Read payload is failed").Write(w)
		return
	}

	var payload ReservationPutPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Parse payload is failed").Write(w)
		return
	}

	var target Reservation
	if err := h.master.SelectOne(&target, "SELECT * FROM reservation WHERE reservation_id = ?", payload.ReservationID); err != nil {
		if err == sql.ErrNoRows {
			NewErrorResponse(http.StatusNotFound, fmt.Sprintf("id=%d is not found", payload.ReservationID)).Write(w)
		} else {
			log.Println(err.Error())
			NewErrorResponse(http.StatusInternalServerError, "Select reservation is failed").Write(w)
		}
		return
	}

	now := time.Now()
	reservation := Reservation{
		ReservationID:       payload.ReservationID,
		UserID:              payload.UserID,
		MenuID:              payload.MenuID,
		ReservationDateTime: payload.ReservationDateTime,
		Created:             target.Created,
		Updated:             now,
	}

	if _, err := h.master.Update(&reservation); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Update Data is failed").Write(w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
