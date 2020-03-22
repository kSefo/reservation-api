package main

import (
	"encoding/json"
	"log"
	"net/http"
)

//　FIXME 共通化したい
// Handler Handlerの中でDBへクエリを発行するため、DB接続の構造体を持つ
// type Handler struct {
// 	master *gorp.DbMap
// }

// NewHandler Handlerの作成関数
// func NewHandler(master *gorp.DbMap) http.Handler {
// 	return &Handler{
// 		master: master,
// 	}
// }

// // HTTPリクエストを受け、ビジネスロジックを実行してレスポンスを返す
// func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	log.Printf("[%s] RemoteAddr=%s\tUserAgent=%s", r.Method, r.RemoteAddr, r.Header.Get("User-Agent"))
// 	switch r.Method {
// 	case "GET":
// 		h.serveGET(w, r)
// 		return
// 	case "POST":
// 		h.servePOST(w, r)
// 		return
// 	case "PUT":
// 		h.servePUT(w, r)
// 		return
// 	default:
// 		NewErrorResponse(http.StatusMethodNotAllowed, fmt.Sprintf("%s is Unsupported method", r.Method)).Write(w)
// 		return
// 	}
// }

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
