package handler

import (
	"db_tpark/repository"
	"github.com/gorilla/mux"
	"net/http"
)

type ForumHandler struct {
	repo repository.Repository
}

func NewForumHandler(repo repository.Repository) *ForumHandler {
	return &ForumHandler{repo: repo}
}

func (h *ForumHandler) InflateRouter(r *mux.Router) {
	r.HandleFunc("/create", h.CreateForum).Methods("POST")
	r.HandleFunc("/{slug}/create", h.CreateThread).Methods("POST")
	r.HandleFunc("/{slug}/details", h.GetForumDetails).Methods("GET")
	r.HandleFunc("/{slug}/threads", h.GetForumThreads).Methods("GET")
	r.HandleFunc("/{slug}/users", h.ForumUsers).Methods("GET")
}

func (h *ForumHandler) CreateForum(w http.ResponseWriter, r *http.Request) {

}

func (h *ForumHandler) CreateThread(w http.ResponseWriter, r *http.Request) {

}

func (h *ForumHandler) GetForumDetails(w http.ResponseWriter, r *http.Request) {

}

func (h *ForumHandler) GetForumThreads(w http.ResponseWriter, r *http.Request) {

}

func (h *ForumHandler) ForumUsers(w http.ResponseWriter, r *http.Request) {

}

func (h *ForumHandler) GetThreadDetails(w http.ResponseWriter, r *http.Request) {

}

