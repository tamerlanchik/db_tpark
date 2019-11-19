package handler

import (
	"db_tpark/repository"
	"db_tpark/structs"
	"fmt"
	"github.com/go-park-mail-ru/2019_2_Next_Level/pkg/HttpTools"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
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
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

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
			resp.
				SetStatus(404).
				SetContent(structs.Error{Message:"Can-t fine user with nick " + forum.User})
			return
		case structs.ErrorDuplicateKey:
			forum, err = h.repo.GetForum(forum.Slug)
			resp.SetStatus(409).SetContent(forum)
		}
		return
	}
	forum, _ = h.repo.GetForum(forum.Slug)
	resp.SetStatus(201).SetContent(forum)
}

func (h *ForumHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

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
	if req.Created == "" {
		req.Created = time.Now().Format(structs.OutTimeFormat)
	}
	thread, err := h.repo.CreateThread(req)
	if err != nil {
		e := err.(structs.InternalError)
		switch e.E{
		case structs.ErrorNoUser:
			resp.
				SetStatus(404).
				SetContent(structs.Error{Message:"Can-t fine user with nick " + req.Author})
			return
		case structs.ErrorNoForum:
			resp.
				SetStatus(404).
				SetContent(structs.Error{Message:"Can-t find forum with slug " + req.Forum})
			return
		case structs.ErrorDuplicateKey:
			thread, err := h.repo.GetThread(req.Slug)
			if err != nil {
				fmt.Println(err)
			}
			resp.
				SetStatus(409).
				SetContent(thread)
			return
		}
	}
	//thread.Slug = ""
	resp.SetStatus(201).SetContent(thread)
}

func (h *ForumHandler) GetForumDetails(w http.ResponseWriter, r *http.Request) {
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	forumSlug, ok := args["slug"]
	if !ok {
		fmt.Println("No such a param: ", "nick")
		return
	}
	var forum structs.Forum
	forum, err := h.repo.GetForum(forumSlug)
	if err != nil {
		resp.
			SetStatus(404).
			SetContent(structs.Error{Message:"Can-t fiтв forum with slug " + forumSlug})
		return
	}
	resp.SetStatus(200).SetContent(forum)

}

func (h *ForumHandler) GetForumThreads(w http.ResponseWriter, r *http.Request) {
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	forumSlug, ok := args["slug"]
	if !ok {
		fmt.Println("No such a param: ", "nick")
		return
	}
	limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 8)

	since := r.FormValue("since")
	desc, err := strconv.ParseBool(r.FormValue("desc"))

	threads, err := h.repo.GetThreads(forumSlug, limit, since, desc)
	if err != nil{
		fmt.Println("Error in GetThreads: ", err, len(threads), forumSlug)
		resp.
			SetStatus(404).SetContent(structs.Error{Message:"Can-t fiтв forum with slug " + forumSlug})
		return
	}
	resp.SetStatus(200).SetContent(threads)
}

func (h *ForumHandler) ForumUsers(w http.ResponseWriter, r *http.Request) {
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

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
	if err != nil || len(users)==0{
		resp.SetStatus(404).SetContent(structs.Error{Message:"Can-t find forum with slug " + forumSlug})
		return
	}
	resp.SetStatus(200).SetContent(users)
}
