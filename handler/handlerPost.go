package handler

import (

	"db_tpark/repository"
	"db_tpark/structs"
	"fmt"
	"github.com/go-park-mail-ru/2019_2_Next_Level/pkg/HttpTools"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	resp := HttpTools.NewResponse(w)
	defer resp.Send()
	debugCounter++;
	fmt.Println("D", debugCounter, time.Now())

	if debugCounter>=35{
		fmt.Println(debugCounter)
	}

	args := mux.Vars(r)
	id, ok := args["id"]
	if !ok {
		fmt.Println("No such a param: ", "nick")
		return
	}
	rel := r.URL.Query()["related"]
	related := make([]string, 0)
	for _, r := range rel {
		related = append(related, strings.Split(r, ",")...)
	}
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		fmt.Println("Wrong param: ", "id")
		return
	}
	posts, err := h.repo.GetPostAccount(idInt, related)
	if err != nil {
		resp.SetStatus(404).SetContent(structs.Error{Message:"Can't find user with id #42"})
		return
	}
	if posts.Post.Parent == posts.Post.Id {
		posts.Post.Parent = 0
	}
	//posts.Post.IsEdited = false;
	resp.SetStatus(200).SetContent(posts)
}

func (h *PostHandler) ChangePost(w http.ResponseWriter, r *http.Request) {
	resp := HttpTools.NewResponse(w)
	defer resp.Send()
	debugCounter++;
	time.Sleep(5*time.Second)
	fmt.Println("A", debugCounter, time.Now())

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
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		fmt.Println("Wrong param: ", "id")
		return
	}
	err = h.repo.EditPost(idInt, post)
	if err != nil {
		resp.SetStatus(404).SetContent(structs.Error{Message:"Can't find user with id #42"})
		return
	}
	post, err = h.repo.GetPost(idInt)
	if err != nil {
		resp.SetStatus(404).SetContent(structs.Error{Message:"Can't find user with id #42"})
		return
	}
	post.IsEdited = true;
	resp.SetStatus(200).SetContent(post)
}