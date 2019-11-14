package handler

import (
	"2019_2_Next_Level/pkg/HttpTools"
	"db_tpark/repository"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type ServiceHandler struct {
	repo repository.Repository
}

func NewServiceHandler(repo repository.Repository) *ServiceHandler {
	return &ServiceHandler{repo: repo}
}

func(h *ServiceHandler) InflateRouter(r *mux.Router) {
	r.HandleFunc("/clear", h.ClearAll).Methods("POST")
	r.HandleFunc("/status", h.GetStatus).Methods("GET")
}

func (h *ServiceHandler) ClearAll(w http.ResponseWriter, r *http.Request) {
	h.repo.ClearAll()
}

func (h *ServiceHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	var resp struct{
		Forum int64 `json:"forum"`
		Post int64 `json:"post"`
		Thread int64 `json:"thread"`
		User int64 `json:"user"`
	}

	data, err := h.repo.GetDBAccount()
	if err != nil {
		fmt.Println(err)
		return
	}
	resp.Forum = data["forum"]
	resp.Post = data["post"]
	resp.Thread = data["thread"]
	resp.User = data["user"]
	HttpTools.BodyFromStruct(w, resp)
}