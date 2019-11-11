package handler

import (
	"db_tpark/repository"
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

}

func (h *ServiceHandler) GetStatus(w http.ResponseWriter, r *http.Request) {

}