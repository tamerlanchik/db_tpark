package handler

import (
	"db_tpark/repository"
	"db_tpark/structs"
	"fmt"
	"github.com/go-park-mail-ru/2019_2_Next_Level/pkg/HttpTools"
	"github.com/gorilla/mux"
	"net/http"
)

type UserHandler struct {
	repo repository.Repository
}

func NewUserHandler(repo repository.Repository) *UserHandler {
	return &UserHandler{repo: repo}
}

func(h *UserHandler) InflateRouter(r *mux.Router) {
	r.HandleFunc("/{nick}/create", h.CreateUser).Methods("POST")
	r.HandleFunc("/{nick}/profile", h.GetUser).Methods("GET")
	r.HandleFunc("/{nick}/profile", h.UpdateUser).Methods("POST")
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	nickname, ok := args["nick"]
	if !ok {
		fmt.Println("No such a param: ", "nick")
		return
	}
	var user structs.User
	err := HttpTools.StructFromBody(*r, &user)
	if err != nil {
		fmt.Println("Cannot parse Createuser body")
		return
	}
	user.Nickname = nickname

	err = h.repo.AddUser(user)
	var ans structs.User
	if err == nil {
		ans = user
		w.WriteHeader(201)
	} else {
		existUser, err := h.repo.GetUser(user.Email, "")
		if err != nil {
			existUser, err = h.repo.GetUser("", user.Nickname)
			if err != nil {
				fmt.Println("user not exist. Cannot create user")
				return
			}
		}
		ans = existUser
		w.WriteHeader(409)
	}
	err = HttpTools.BodyFromStruct(w, ans)
	if err != nil {
		fmt.Println("Cannot write to body")
	}
	return
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	nickname, ok := args["nick"]
	if !ok {
		fmt.Println("No such a param: ", "nick")
		return
	}
	user, err := h.repo.GetUser("", nickname)
	if err != nil {
		fmt.Println(err)
		err = HttpTools.BodyFromStruct(w, struct{
			Message string `json:"message"`
		}{Message:"Can-t find user with nickname " + nickname})
		if err != nil {
			fmt.Println(err)
		}
		w.WriteHeader(404)
		return
	}
	err = HttpTools.BodyFromStruct(w, user)
	if err!=nil {
		fmt.Println(err)
	}
	w.WriteHeader(200)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	nickname, ok := args["nick"]
	if !ok {
		fmt.Println("No such a param: ", "nick")
		return
	}
	var user structs.User
	err := HttpTools.StructFromBody(*r, &user)
	if err != nil {
		fmt.Println("Cannot parse Updateeuser body")
		return
	}
	user.Nickname = nickname

	err = h.repo.EditUser(user)
	if err != nil {
		w.WriteHeader(404)
		err = HttpTools.BodyFromStruct(w, struct{
			Message string `json:"message"`
		}{Message:"Can-t find user with nickname " + nickname})
		if err != nil {
			fmt.Println(err)
		}
	}
	err = HttpTools.BodyFromStruct(w, user)
	if err != nil {
		fmt.Println("Cannot write to body")
	}
	return
}