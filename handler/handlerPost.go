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
	r.HandleFunc("/{id}/details", h.GetPostDetails).Methods("GET")
	r.HandleFunc("/{id}/details", h.ChangePost).Methods("POST")
}

func (h *PostHandler) GetPostDetails(w http.ResponseWriter, r *http.Request) {
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

func (h *PostHandler) ChangePost(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	id, ok := args["id"]
	if !ok {
		fmt.Println("No such a param: ", "nick")
		return
	}

	var post structs.Post
	err := HttpTools.StructFromBody(*r, &post)
	if err != nil {
		fmt.Println("No such a param: ", "post")
		return
	}
	idInt, err := strconv.ParseInt(id, 10, 8)
	if err != nil {
		fmt.Println("Wrong param: ", "id")
		return
	}
	err = h.repo.EditPost(idInt, post)
	if err != nil {
		HttpTools.BodyFromStruct(w, structs.Error{Message:"Can't find user with id #42"})
		w.WriteHeader(404)
		return
	}
	post, err = h.repo.GetPost(idInt)
	if err != nil {
		HttpTools.BodyFromStruct(w, structs.Error{Message:"Can't find user with id #42"})
		w.WriteHeader(404)
		return
	}
	HttpTools.BodyFromStruct(w, post)
	w.WriteHeader(200)
}