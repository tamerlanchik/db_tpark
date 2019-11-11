package handler

import (
	"db_tpark/repository"
	"github.com/gorilla/mux"
	"net/http"
)

type PostHandler struct {
	repo repository.Repository
}

func NewPostHandler(repo repository.Repository) *PostHandler {
	return &PostHandler{repo: repo}
}

func (h *PostHandler) InflateRouter(r *mux.Router) {
	r.HandleFunc("/{id}/details", h.GetThreadDetails).Methods("GET")
	r.HandleFunc("/{id{/details", h.ChangeThread).Methods("GET")
}

func (h *PostHandler) GetThreadDetails(w http.ResponseWriter, r *http.Request) {

}

func (h *PostHandler) ChangeThread(w http.ResponseWriter, r *http.Request) {

}