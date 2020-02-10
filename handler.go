package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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

type errorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// NewErrorResponse エラーを返す
func NewErrorResponse(status int, message string) *errorResponse {
	return &errorResponse{
		Status:  status,
		Message: message,
	}
}

func (e *errorResponse) Write(w http.ResponseWriter) {
	data, err := json.Marshal(e)
	if err != nil {
		log.Println("marshal error json is failed")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(e.Status)
	w.Write(data)
}

func (h ReservationHandler) serveGET(w http.ResponseWriter, r *http.Request) {
	reservationDate := r.URL.Query().Get("reservationDate")
	if reservationDate == "" {
		// 現在日付を設定
		reservationDate = time.Now().Format("2006-01-02")
	}
	// 6日後を設定
	reservationDateTime, _ := time.Parse("2006-01-02", reservationDate)
	reservationDateAfter6days := reservationDateTime.AddDate(0, 0, 6).Format("2006-01-02")

	result, err := h.master.Select(Reservation{}, strings.Join([]string{
		"SELECT * FROM reservation",
		"WHERE reservation_datetime BETWEEN ? AND ?",
		"ORDER BY reservation_datetime"}, " "), reservationDate+" 00:00:00", reservationDateAfter6days+" 23:59:59")
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

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

// ReservationPostPayload （新規登録用）予約テーブルの構造体
type ReservationPostPayload struct {
	ReservationDatetime time.Time `json:"reservation_datetime"`
	ReservationName     string    `json:"reservation_name"`
	Request             string    `json:"request"`
	Status              string    `json:"status"`
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
		ReservationDatetime: payload.ReservationDatetime,
		ReservationName:     payload.ReservationName,
		Request:             payload.Request,
		Status:              "×",
		Created:             now,
		Updated:             now,
	}

	if err := h.master.Insert(&reservation); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Insert Data is failed").Write(w)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
}

// ReservationPutPayload （更新用）予約テーブルの構造体
type ReservationPutPayload struct {
	ID                  uint      `json:"id"`
	ReservationDatetime time.Time `json:"reservation_datetime"`
	ReservationName     string    `json:"reservation_name"`
	Request             string    `json:"request"`
	Status              string    `json:"status"`
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
	if err := h.master.SelectOne(&target, "SELECT * FROM reservation WHERE id = ?", payload.ID); err != nil {
		if err == sql.ErrNoRows {
			NewErrorResponse(http.StatusNotFound, fmt.Sprintf("id=%d is not found", payload.ID)).Write(w)
		} else {
			log.Println(err.Error())
			NewErrorResponse(http.StatusInternalServerError, "Select reservation is failed").Write(w)
		}
		return
	}

	now := time.Now()
	reservation := Reservation{
		ID:                  payload.ID,
		ReservationDatetime: payload.ReservationDatetime,
		ReservationName:     payload.ReservationName,
		Request:             payload.Request,
		Status:              payload.Status,
		Created:             target.Created,
		Updated:             now,
	}

	if _, err := h.master.Update(&reservation); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Update Data is failed").Write(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
