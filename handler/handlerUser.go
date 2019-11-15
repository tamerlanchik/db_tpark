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
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

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
	if err == nil {
		resp.SetStatus(201).SetContent(user)
		return
	} else {
		var ans []structs.User

		existUserByEmail, err := h.repo.GetUser(user.Email, "")
		if err == nil {
			ans = append(ans, existUserByEmail)
		}

		existUserByNick, err := h.repo.GetUser("", user.Nickname)
		if err == nil && existUserByNick.Email!=existUserByEmail.Email{
			ans = append(ans, existUserByNick)
		}
		resp.SetStatus(409).SetContent(ans)
	}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	nickname, ok := args["nick"]
	if !ok {
		fmt.Println("No such a param: ", "nick")
		return
	}
	user, err := h.repo.GetUser("", nickname)
	if err != nil {
		fmt.Println(err)
		resp.
			SetStatus(404).
			SetContent(struct{
					Message string `json:"message"`
				}{Message:"Can-t find user with nickname " + nickname})

		return
	}
	resp.SetStatus(200).SetContent(user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

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
		resp.
			SetStatus(404).
			SetContent(struct{
					Message string `json:"message"`
				}{Message:"Can-t find user with nickname " + nickname})
	}
	resp.SetStatus(200).SetError(user)
	return
}