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
	var forum structs.Forum
	err := HttpTools.StructFromBody(*r, &forum)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = h.repo.CreateForum(forum.Slug, forum.Title, forum.User)
	if err != nil {
		e := err.(structs.InternalError)
		switch e.E{
		case structs.ErrorNoUser:
			HttpTools.BodyFromStruct(w, structs.Error{Message:"Can-t fine user with nick " + forum.User})
			w.WriteHeader(404)
			return
		case structs.ErrorDuplicateKey:
			forum, err = h.repo.GetForum(forum.Slug)
			w.WriteHeader(409)
			HttpTools.BodyFromStruct(w, forum)
		}
	}else{
		w.WriteHeader(201)
		HttpTools.BodyFromStruct(w, forum)
	}
}

func (h *ForumHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	forumSlug, ok := args["slug"]
	if !ok {
		fmt.Println("No such a param: ", "nick")
		return
	}
	var req structs.Thread
	err := HttpTools.StructFromBody(*r, &req)
	if err != nil {
		fmt.Println("Wrror body in CreateThread")
		return
	}
	req.Forum = forumSlug
	thread, err := h.repo.CreateThread(req)
	if err != nil {
		e := err.(structs.InternalError)
		switch e.E{
		case structs.ErrorNoUser:
			HttpTools.BodyFromStruct(w, structs.Error{Message:"Can-t fine user with nick " + req.Author})
			w.WriteHeader(404)
			return
		case structs.ErrorNoForum:
			HttpTools.BodyFromStruct(w, structs.Error{Message:"Can-t find forum with slug " + req.Forum})
			w.WriteHeader(404)
			return
		case structs.ErrorDuplicateKey:
			thread, err := h.repo.GetThread(req.Slug)
			if err != nil {
				fmt.Println(err)
			}
			w.WriteHeader(409)
			HttpTools.BodyFromStruct(w, thread)
		}
	}
	w.WriteHeader(201)
	HttpTools.BodyFromStruct(w, thread)
}

func (h *ForumHandler) GetForumDetails(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	forumSlug, ok := args["slug"]
	if !ok {
		fmt.Println("No such a param: ", "nick")
		return
	}
	var forum structs.Forum
	forum, err := h.repo.GetForum(forumSlug)
	if err != nil {
		HttpTools.BodyFromStruct(w, structs.Error{Message:"Can-t fiтв forum with slug " + forumSlug})
		w.WriteHeader(404)
		return
	}
	HttpTools.BodyFromStruct(w, forum)
	w.WriteHeader(200)

}

func (h *ForumHandler) GetForumThreads(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	forumSlug, ok := args["slug"]
	if !ok {
		fmt.Println("No such a param: ", "nick")
		return
	}
	limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 8)
	if err != nil {
		return
	}
	since := r.FormValue("since")
	desc, err := strconv.ParseBool(r.FormValue("desc"))
	if err != nil {
		return
	}

	threads, err := h.repo.GetThreads(forumSlug, limit, since, desc)
	if err != nil {
		HttpTools.BodyFromStruct(w, structs.Error{Message:"Can-t fiтв forum with slug " + forumSlug})
		w.WriteHeader(404)
		return
	}
	HttpTools.BodyFromStruct(w, threads)
	w.WriteHeader(200)
}

func (h *ForumHandler) ForumUsers(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	forumSlug, ok := args["slug"]
	if !ok {
		fmt.Println("No such a param: ", "slug")
		return
	}
	limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 8)
	if err != nil {
		return
	}
	since := r.FormValue("since")
	desc, err := strconv.ParseBool(r.FormValue("desc"))
	if err != nil {
		return
	}

	users, err := h.repo.GetUsers(forumSlug, limit, since, desc)
	if err != nil {
		HttpTools.BodyFromStruct(w, structs.Error{Message:"Can-t fiтв forum with slug " + forumSlug})
		w.WriteHeader(404)
		return
	}
	HttpTools.BodyFromStruct(w, users)
	w.WriteHeader(200)
}

func (h *ForumHandler) GetThreadDetails(w http.ResponseWriter, r *http.Request) {

}

