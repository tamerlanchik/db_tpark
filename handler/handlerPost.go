package handler

import (
	"db_tpark/buildmode"
	"db_tpark/repository"
	"db_tpark/structs"
	"time"

	//"fmt"
	"db_tpark/pkg/HttpTools"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
	//"time"
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
	//tic := time.Now()
	//defer timeLogger.Write("/post/getDetails", tic)
	//buildmode.Log.Println("GetPostDetails", GetPostCounter)
	//if GetPostCounter>=18{
	//	//buildmode.Log.Println("GetPostDetails")
	//}
	resp := HttpTools.NewResponse(w)
	defer resp.Send()
	debugCounter++;
	//buildmode.Log.Println("D", debugCounter, time.Now())


	args := mux.Vars(r)
	id, ok := args["id"]
	if !ok {
		buildmode.Log.Println("No such a param: ", "nick")
		return
	}
	rel := r.URL.Query()["related"]
	related := make([]string, 0)
	for _, r := range rel {
		related = append(related, strings.Split(r, ",")...)
	}
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		buildmode.Log.Println("Wrong param: ", "id")
		return
	}

	posts, err := h.repo.GetPostAccount(idInt, related)
	//buildmode.Log.Println("After GetPostDetails")
	//PrintMemUsage()
	if err != nil {
		buildmode.Log.Println("Error in GetPostDetails: ", err)
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
	tic := time.Now()
	defer timeLogger.Write("/post/change", tic)
	//EditPostCounter++
	//buildmode.Log.Println("Edit post", EditPostCounter)
	//if EditPostCounter>=3 {
	//	//buildmode.Log.Println("Edit post", EditPostCounter)
	//}
	resp := HttpTools.NewResponse(w)
	defer resp.Send()
	debugCounter++;
	//time.Sleep(5*time.Second)
	//buildmode.Log.Println("A", debugCounter, time.Now())

	args := mux.Vars(r)
	id, ok := args["id"]
	if !ok {
		buildmode.Log.Println("No such a param: ", "nick")
		return
	}

	var post structs.Post
	err := HttpTools.StructFromBody(*r, &post)
	if err != nil {
		buildmode.Log.Println("No such a param: ", "post")
		return
	}
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		buildmode.Log.Println("Wrong param: ", "id")
		return
	}
	err = h.repo.EditPost(idInt, post)
	if err != nil {
		buildmode.Log.Println("Error in ChangePost-EdiPost: ", err)
		resp.SetStatus(404).SetContent(structs.Error{Message:"Can't find user with id #42"})
		return
	}
	post, err = h.repo.GetPost(idInt)
	if err != nil {
		buildmode.Log.Println("Error in ChangePost-GetPost: ", err)
		resp.SetStatus(404).SetContent(structs.Error{Message:"Can't find user with id #42"})
		return
	}
	resp.SetStatus(200).SetContent(post)
}