package handler

import (
	"db_tpark/repository"
	"github.com/gorilla/mux"
	"net/http"
)

type ThreadHandler struct {
	repo repository.Repository
}

func NewThreadHandler(repo repository.Repository) *ThreadHandler {
	return &ThreadHandler{repo: repo}
}

func(h *ThreadHandler) InflateRouter(r *mux.Router) {
	r.HandleFunc("/{id}/create", h.CreateThread).Methods("POST")
	r.HandleFunc("/{id}/details", h.GetDetails).Methods("GET")
	r.HandleFunc("/{id}/details", h.UpdateThread).Methods("POST")
	r.HandleFunc("/{id}/posts", h.GetPosts).Methods("GET")
	r.HandleFunc("/{id}/vote", h.Vote).Methods("POST")

}

func (h *ThreadHandler) CreateThread(w http.ResponseWriter, r *http.Request) {

}

func (h *ThreadHandler) GetDetails(w http.ResponseWriter, r *http.Request) {

}

func (h *ThreadHandler) UpdateThread(w http.ResponseWriter, r *http.Request) {

}

func (h *ThreadHandler) Vote(w http.ResponseWriter, r *http.Request) {

}

func (h *ThreadHandler) GetPosts(w http.ResponseWriter, r *http.Request) {

}

