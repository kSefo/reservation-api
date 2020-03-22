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

// SalesDayHandler Handlerの中でDBへクエリを発行するため、DB接続の構造体を持つ
type SalesDayHandler struct {
	master *gorp.DbMap
}

// NewSalesDayHandler Handlerの作成関数
func NewSalesDayHandler(master *gorp.DbMap) http.Handler {
	return &SalesDayHandler{
		master: master,
	}
}

// HTTPリクエストを受け、ビジネスロジックを実行してレスポンスを返す
func (h SalesDayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h SalesDayHandler) serveGET(w http.ResponseWriter, r *http.Request) {
	result, err := h.master.Select(SalesDay{}, "SELECT * FROM sales_day")
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Execute Query is failed").Write(w)
		return
	}

	salesDayItems := make([]*SalesDay, 0)

	for _, e := range result {
		salesDay := e.(*SalesDay)
		salesDayItems = append(salesDayItems, salesDay)
	}

	data, err := json.Marshal(salesDayItems)
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Marshal JSON is failed").Write(w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

// SalesDayPostPayload （新規登録用）ユーザテーブルの構造体
type SalesDayPostPayload struct {
	SalesDay  string `json:"sales_day"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Holiday   bool   `json:"holiday"`
}

func (h SalesDayHandler) servePOST(w http.ResponseWriter, r *http.Request) {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Read payload is failed").Write(w)
		return
	}

	var payload SalesDayPostPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Parse payload is failed").Write(w)
		return
	}

	now := time.Now()
	salesDay := SalesDay{
		SalesDay:  payload.SalesDay,
		StartTime: payload.StartTime,
		EndTime:   payload.EndTime,
		Holiday:   payload.Holiday,
		Created:   now,
		Updated:   now,
	}

	if err := h.master.Insert(&salesDay); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Insert Data is failed").Write(w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
}

// SalesDayPutPayload （更新用）ユーザテーブルの構造体
type SalesDayPutPayload struct {
	SalesDay  string `json:"sales_day"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Holiday   bool   `json:"holiday"`
}

func (h SalesDayHandler) servePUT(w http.ResponseWriter, r *http.Request) {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Read payload is failed").Write(w)
		return
	}

	var payload SalesDayPutPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Parse payload is failed").Write(w)
		return
	}

	var target SalesDay
	if err := h.master.SelectOne(&target, "SELECT * FROM sales_day WHERE sales_day = ?", payload.SalesDay); err != nil {
		if err == sql.ErrNoRows {
			NewErrorResponse(http.StatusNotFound, fmt.Sprintf("sales_day=%s is not found", payload.SalesDay)).Write(w)
		} else {
			log.Println(err.Error())
			NewErrorResponse(http.StatusInternalServerError, "Select sales_day is failed").Write(w)
		}
		return
	}

	now := time.Now()
	salesDay := SalesDay{
		SalesDay:  payload.SalesDay,
		StartTime: payload.StartTime,
		EndTime:   payload.EndTime,
		Holiday:   payload.Holiday,
		Created:   target.Created,
		Updated:   now,
	}

	if _, err := h.master.Update(&salesDay); err != nil {
		log.Println(err.Error())
		NewErrorResponse(http.StatusInternalServerError, "Update Data is failed").Write(w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
