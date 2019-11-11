package handler

import (
	"db_tpark/repository"
	"github.com/gorilla/mux"
	"net/http"
)

type UserHandler struct {
	repo repository.Repository
}

func NewUserHandler(repo repository.Repository) *UserHandler {
	return &UserHandler{repo: repo}
}

func(h *UserHandler) InflateRouter(r *mux.Router) {
	r.HandleFunc("/{nick}/create", h.CreateUser).Methods("POST")
	r.HandleFunc("/{nick}/profile", h.GetUser).Methods("GET")
	r.HandleFunc("/{nick}/profile", h.UpdateUser).Methods("POST")
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {

}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {

}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {

}