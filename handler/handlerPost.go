package handler

import (

	"db_tpark/repository"
	"db_tpark/structs"
	"fmt"
	"db_tpark/pkg/HttpTools"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PostHandler struct {
	repo repository.Repository
}

var GetPostCounter int
var EditPostCounter int

func NewPostHandler(repo repository.Repository) *PostHandler {
	return &PostHandler{repo: repo}
}

func (h *PostHandler) InflateRouter(r *mux.Router) {
	r.HandleFunc("/{id}/details", h.GetPostDetails).Methods("GET")
	r.HandleFunc("/{id}/details", h.ChangePost).Methods("POST")
}

func (h *PostHandler) GetPostDetails(w http.ResponseWriter, r *http.Request) {
	GetPostCounter++
	fmt.Println("GetPostDetails", GetPostCounter)
	if GetPostCounter>=18{
		fmt.Println("GetPostDetails")
	}
	resp := HttpTools.NewResponse(w)
	defer resp.Send()
	debugCounter++;
	fmt.Println("D", debugCounter, time.Now())


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
	EditPostCounter++
	fmt.Println("Edit post", EditPostCounter)
	if EditPostCounter>=3 {
		fmt.Println("Edit post", EditPostCounter)
	}
	resp := HttpTools.NewResponse(w)
	defer resp.Send()
	debugCounter++;
	//time.Sleep(5*time.Second)
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
	resp.SetStatus(200).SetContent(post)
}