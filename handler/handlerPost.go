package handler

import (
	"2019_2_Next_Level/pkg/HttpTools"
	"db_tpark/repository"
	"db_tpark/structs"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
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
	args := mux.Vars(r)
	id, ok := args["id"]
	if !ok {
		fmt.Println("No such a param: ", "nick")
		return
	}
	related:= r.URL.Query()["related"]
	idInt, err := strconv.ParseInt(id, 10, 8)
	if err != nil {
		fmt.Println("Wrong param: ", "id")
		return
	}
	posts, err := h.repo.GetPostAccount(idInt, related)
	if err != nil {
		HttpTools.BodyFromStruct(w, structs.Error{Message:"Can't find user with id #42"})
		w.WriteHeader(404)
		return
	}
	HttpTools.BodyFromStruct(w, posts)
	w.WriteHeader(200)
}

func (h *PostHandler) ChangeThread(w http.ResponseWriter, r *http.Request) {

}