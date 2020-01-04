package handler

import (
	"db_tpark/buildmode"
	"db_tpark/repository"
	"db_tpark/structs"
	"fmt"
	"runtime"

	//"fmt"
	"db_tpark/pkg/HttpTools"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

type ForumHandler struct {
	repo repository.Repository
}

func init() {
	if timeLogger == nil {
		timeLogger = &MockLogger{}
		timeLogger.Init()
	}
}


func PrintMetrics() {
	d := timeLogger
	for path, times := range d.Data() {
		buildmode.Log.Println(path)
		list := make([]int64, 0)
		for _, elem := range times{
			if elem > 5 {
				list = append(list, elem)
			}
		}
		fmt.Println(list)
		fmt.Printf("Avr: %d\n", func() int64 {
			var res int64
			for _, val := range times {
				res += val
			}
			res = res/int64(len(times))
			return res
		}())
	}
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", m.Alloc)
	fmt.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc)
	fmt.Printf("\tSys = %v MiB", m.Sys)
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

var timeLogger Logger


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
	tic := time.Now()
	defer timeLogger.Write("/forum/create", tic)
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	var forum structs.Forum
	err := HttpTools.StructFromBody(*r, &forum)
	if err != nil {
		buildmode.Log.Println(err)
		return
	}
	err = h.repo.CreateForum(forum.Slug, forum.Title, forum.User)
	if err != nil {
		buildmode.Log.Println("Error in Create Forum: ", err)
		e := err.(structs.InternalError)
		switch e.E{
		case structs.ErrorNoUser:
			//buildmode.Log.Println("Error in Create Forum: ", err)
			resp.
				SetStatus(404).
				SetContent(structs.Error{Message:"Can-t fine user with nick " + forum.User})
			return
		case structs.ErrorDuplicateKey:
			forum, err = h.repo.GetForum(forum.Slug)
			resp.SetStatus(409).SetContent(forum)
		}
		//buildmode.Log.Println("Error in Create Forum: ", err)
		return
	}
	forum, _ = h.repo.GetForum(forum.Slug)
	resp.SetStatus(201).SetContent(forum)
}

func (h *ForumHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	tic := time.Now()
	defer timeLogger.Write("/thread/create", tic)
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	forumSlug, ok := args["slug"]
	if !ok {
		buildmode.Log.Println("No such a param: ", "nick")
		return
	}
	var req structs.Thread
	err := HttpTools.StructFromBody(*r, &req)
	if err != nil {
		buildmode.Log.Println("Wrror body in CreateThread")
		return
	}
	req.Forum = forumSlug
	if req.Created == "" {
		req.Created = time.Now().Format(structs.OutTimeFormat)
	}
	thread, err := h.repo.CreateThread(req)
	if err != nil {
		buildmode.Log.Println("Error in CreateThread: ", err)
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
				//buildmode.Log.Println(err)
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
	tic := time.Now()
	defer timeLogger.Write("/forum/details", tic)
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	forumSlug, ok := args["slug"]
	if !ok {
		//buildmode.Log.Println("No such a param: ", "nick")
		return
	}
	var forum structs.Forum
	forum, err := h.repo.GetForum(forumSlug)
	if err != nil {
		buildmode.Log.Println("Error in GetForumDetails: ", err)
		resp.
			SetStatus(404).
			SetContent(structs.Error{Message:"Can-t fiтв forum with slug " + forumSlug})
		return
	}
	resp.SetStatus(200).SetContent(forum)

}

func (h *ForumHandler) GetForumThreads(w http.ResponseWriter, r *http.Request) {
	tic := time.Now()
	defer timeLogger.Write("/forum/threads", tic)
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	forumSlug, ok := args["slug"]
	if !ok {
		//buildmode.Log.Println("No such a param: ", "nick")
		return
	}
	limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 64)

	since := r.FormValue("since")
	desc, err := strconv.ParseBool(r.FormValue("desc"))

	threads, err := h.repo.GetThreads(forumSlug, limit, since, desc)
	if err != nil{
		buildmode.Log.Println("Error in GetThreads: ", err)
		resp.
			SetStatus(404).SetContent(structs.Error{Message:"Can-t fiтв forum with slug " + forumSlug})
		return
	}
	resp.SetStatus(200).SetContent(threads)
}

func (h *ForumHandler) ForumUsers(w http.ResponseWriter, r *http.Request) {
	tic := time.Now()
	defer timeLogger.Write("/forum/users", tic)
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	forumSlug, ok := args["slug"]
	if !ok {
		buildmode.Log.Println("No such a param: ", "slug")
		return
	}
	limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 64)
	//if err != nil {
	//	return
	//}
	since := r.FormValue("since")
	desc, err := strconv.ParseBool(r.FormValue("desc"))
	if err != nil {
		desc = false
	}

	forum, err := h.repo.GetForum(forumSlug)
	if err != nil || forum.Slug==""{
		buildmode.Log.Println("Error in ForumUsers-GetForum: ", err)
		resp.SetStatus(404).SetContent(structs.Error{Message:"Can-t find forum with slug " + forumSlug})
		return
	}

	users, err := h.repo.GetUsers(forumSlug, limit, since, desc)
	if err != nil{
		buildmode.Log.Println("Error in ForumUsers-GetUsers: ", err)
		resp.SetStatus(404).SetContent(structs.Error{Message:"Can-t find forum with slug " + forumSlug})
		return
	}
	resp.SetStatus(200).SetContent(users)
}
