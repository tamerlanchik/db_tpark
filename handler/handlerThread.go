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
	args := mux.Vars(r)
	id, ok := args["id"]
	threadId, err := strconv.ParseInt(id, 10, 8)
	if !ok || err != nil {
		fmt.Println("No such a param: ", "nick")
		return
	}
	var post []structs.Post
	err = HttpTools.StructFromBody(*r, &post)
	if err != nil {
		fmt.Println("Error struct got")
		return
	}
	for i:=range post{
		post[i].Thread = int32(threadId)
	}
	post, err = h.repo.CreatePost(post)
	if err != nil {
		switch err.(structs.InternalError).E{
		case structs.ErrorNoThread:
			w.WriteHeader(404)
			break
		case structs.ErrorNoParent:
			w.WriteHeader(409)
			break
		}
		HttpTools.BodyFromStruct(w, structs.Error{Message:"Can't find user with id #42\n"})
		return
	}
	HttpTools.BodyFromStruct(w, post)
	w.WriteHeader(201)
}

func (h *ThreadHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	id, ok := args["id"]
	if !ok {
		fmt.Println("No such a param: ", "id")
		return
	}
	threadId, err := strconv.ParseInt(id, 10, 8)

	var thread structs.Thread
	if err == nil {
		thread, err = h.repo.GetThreadById(threadId)
	} else {
		thread, err = h.repo.GetThread(id)
	}

	if err != nil {
		w.WriteHeader(404)
		HttpTools.BodyFromStruct(w, structs.Error{Message:"Can't find user with id #42\n"})
		return
	}

	HttpTools.BodyFromStruct(w, thread)
	w.WriteHeader(200)
}

func (h *ThreadHandler) UpdateThread(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	id, ok := args["id"]
	if !ok {
		fmt.Println("No such a param: ", "id")
		return
	}

	var thread structs.Thread

	err := HttpTools.StructFromBody(*r, &thread)
	if err != nil {
		fmt.Println("Invalid body")
		return
	}

	threadId, err := strconv.ParseInt(id, 10, 8)
	if err == nil {
		thread.Id = int32(threadId)
	} else {
		thread.Slug= id
	}

	err = h.repo.EditThread(thread)

	if err != nil {
		w.WriteHeader(404)
		HttpTools.BodyFromStruct(w, structs.Error{Message:"Can't find user with id #42\n"})
		return
	}

	if thread.Slug != ""{
		thread, err = h.repo.GetThread(thread.Slug)
	}else{
		thread, err = h.repo.GetThreadById(int64(thread.Id))
	}
	if err != nil {
		w.WriteHeader(404)
		HttpTools.BodyFromStruct(w, structs.Error{Message:"Can't find user with id #42\n"})
		return
	}

	HttpTools.BodyFromStruct(w, thread)
	w.WriteHeader(200)
}

func (h *ThreadHandler) Vote(w http.ResponseWriter, r *http.Request) {

}

func (h *ThreadHandler) GetPosts(w http.ResponseWriter, r *http.Request) {

}

