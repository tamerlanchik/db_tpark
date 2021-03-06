package handler

import (
	"db_tpark/buildmode"
	"db_tpark/repository"
	"db_tpark/structs"
	//"fmt"
	"db_tpark/pkg/HttpTools"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	//"time"
)

type ThreadHandler struct {
	repo repository.Repository
}
var counter int
func NewThreadHandler(repo repository.Repository) *ThreadHandler {
	return &ThreadHandler{repo: repo}
}

var debugCounter int

func(h *ThreadHandler) InflateRouter(r *mux.Router) {
	r.HandleFunc("/{id}/create", h.CreatePost).Methods("POST")
	r.HandleFunc("/{id}/details", h.GetDetails).Methods("GET")
	r.HandleFunc("/{id}/details", h.UpdateThread).Methods("POST")
	r.HandleFunc("/{id}/posts", h.GetPosts).Methods("GET")
	r.HandleFunc("/{id}/vote", h.Vote).Methods("POST")

}

func (h *ThreadHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	idOrSlug, ok := args["id"]
	if !ok {
		buildmode.Log.Println("No such a param: ", "nick")
		return
	}

	var threadId interface{}

	// slug or id
	if temp, errId := strconv.ParseInt(idOrSlug, 10, 64); errId==nil {
		threadId = temp
	} else {
		threadId = idOrSlug
	}

	var post []structs.Post
	err := HttpTools.StructFromBody(*r, &post)
	if err != nil {
		buildmode.Log.Println("Error struct got")
		return
	}
	post, err = h.repo.CreatePost(threadId, post)
	if err != nil {
		switch err.(structs.InternalError).E{
		case structs.ErrorNoThread:
			resp.SetStatus(404)
			break
		case structs.ErrorNoParent:
			resp.SetStatus(409)
			break
		default:
			resp.SetStatus(409)
		}
		resp.SetContent(structs.Error{Message:"Can't find user with id #42\n" + err.(structs.InternalError).Explain})
		return
	}
	resp.SetStatus(201).SetContent(post)
}

func (h *ThreadHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	id, ok := args["id"]
	if !ok {
		buildmode.Log.Println("No such a param: ", "id")
		return
	}

	threadId, err := strconv.ParseInt(id, 10, 64)
	var thread structs.Thread
	if err == nil {
		thread, err = h.repo.GetThread(threadId)
	} else {
		thread, err = h.repo.GetThread(id)
	}

	if err != nil {
		resp.SetStatus(404).SetContent(structs.Error{Message:"Can't find user with id #42\n"})
		return
	}

	resp.SetStatus(200).SetContent(thread)
}

func (h *ThreadHandler) UpdateThread(w http.ResponseWriter, r *http.Request) {
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	id, ok := args["id"]
	if !ok {
		buildmode.Log.Println("No such a param: ", "id")
		return
	}

	var thread structs.Thread

	err := HttpTools.StructFromBody(*r, &thread)
	if err != nil {
		buildmode.Log.Println("Invalid body")
		return
	}

	threadId, err := strconv.ParseInt(id, 10, 64)
	if err == nil {
		thread.Id = int32(threadId)
	} else {
		thread.Slug= id
	}

	err = h.repo.EditThread(thread)
	if err != nil {
		resp.SetStatus(404).SetContent(structs.Error{Message:"Can't find user with id #42\n"})
		return
	}

	if thread.Slug != ""{
		thread, err = h.repo.GetThread(thread.Slug)
	}else{
		thread, err = h.repo.GetThread(int64(thread.Id))
	}
	if err != nil {
		resp.SetStatus(404).SetContent(structs.Error{Message:"Can't find user with id #42\n"})
		return
	}

	resp.SetStatus(200).SetContent(thread)
}

func (h *ThreadHandler) Vote(w http.ResponseWriter, r *http.Request) {
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	idString, ok := args["id"]
	var id interface{}
	if !ok {
		buildmode.Log.Println("Error Vote: No such a param: ", "id")
		return
	}
	var err error
	if id, err = strconv.ParseInt(idString, 10, 64); err != nil {
		id = idString
	}

	var req struct {
		Nickname string `json:"nickname"`
		Voice int `json:"voice"`
	}

	err = HttpTools.StructFromBody(*r, &req)
	if err != nil {
		buildmode.Log.Println("Error Vote: Invalid body")
		return
	}

	err = h.repo.VoteThread(id, req.Nickname, req.Voice)
	if err != nil {
		buildmode.Log.Println("Error in Vote: ", err)
		resp.SetStatus(404).SetContent(structs.Error{Message:"Can't find user with id #42\n"})
		return
	}
	thread, err := h.repo.GetThread(id)
	if err != nil {
		buildmode.Log.Println("Error in Vote: cannot get thread: ", err)
		resp.SetStatus(404).SetContent(structs.Error{Message:"Can't find user with id #42\n"})
		return
	}
	resp.SetStatus(200).SetContent(thread)
}

func (h *ThreadHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	threadKey, ok := args["id"]
	if !ok {
		buildmode.Log.Println("No such a param: ", "nick")
		return
	}
	limit, _ := strconv.ParseInt(r.FormValue("limit"), 10, 64)


	since := r.FormValue("since")
	sort := r.FormValue("sort")
	if sort=="" {
		sort = "flat"
	}
	desc, _ := strconv.ParseBool(r.FormValue("desc"))

	posts, err := h.repo.GetPosts(threadKey, limit, since, sort, desc)
	if err != nil{
		buildmode.Log.Println("Error in GetPosts: ", err)
		resp.
			SetStatus(404).SetContent(structs.Error{Message:"Can-t fiтв forum with slug " + threadKey})
		return
	}
	resp.SetStatus(200).SetContent(posts)
}


