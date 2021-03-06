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
	tic := time.Now()
	defer timeLogger.Write("/user/create", tic)
	//buildmode.Log.Println("CreateUser: prestart")
	resp := HttpTools.NewResponse(w)
	//buildmode.Log.Println("CreateUser: resp", resp)
	defer resp.Send()
	//buildmode.Log.Println("CreateUser: start")

	args := mux.Vars(r)
	nickname, ok := args["nick"]
	if !ok {
		buildmode.Log.Println("No such a param: ", "nick")
		return
	}
	var user structs.User
	err := HttpTools.StructFromBody(*r, &user)
	if err != nil {
		buildmode.Log.Println("Cannot parse Createuser body")
		return
	}
	user.Nickname = nickname

	//buildmode.Log.Println("CreateUser: before-repo")
	err = h.repo.AddUser(user)
	//buildmode.Log.Println("CreateUser: after-repo")
	if err == nil {
		resp.SetStatus(201).SetContent(user)
		return
	} else {
		var ans []structs.User
		buildmode.Log.Println("Error in CreateUser-AddUser: ", err)
		existUserByEmail, err := h.repo.GetUser(user.Email, "")
		if err == nil {
			buildmode.Log.Println("Error in CreateUser-GetUser: ", err)
			ans = append(ans, existUserByEmail)
		}

		existUserByNick, err := h.repo.GetUser("", user.Nickname)
		if err == nil && existUserByNick.Email!=existUserByEmail.Email{
			ans = append(ans, existUserByNick)
		}
		//buildmode.Log.Println("Error in Create user: ", err)
		resp.SetStatus(409).SetContent(ans)
	}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	tic := time.Now()
	defer timeLogger.Write("/user/get", tic)
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	nickname, ok := args["nick"]
	if !ok {
		//buildmode.Log.Println("No such a param: ", "nick")
		return
	}
	user, err := h.repo.GetUser("", nickname)
	if err != nil {
		buildmode.Log.Println("Error in GetUser: ", err)
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
	tic := time.Now()
	defer timeLogger.Write("/user/update", tic)
	resp := HttpTools.NewResponse(w)
	defer resp.Send()

	args := mux.Vars(r)
	nickname, ok := args["nick"]
	if !ok {
		buildmode.Log.Println("No such a param: ", "nick")
		return
	}
	var user structs.User
	err := HttpTools.StructFromBody(*r, &user)
	if err != nil {
		buildmode.Log.Println("Cannot parse Updateeuser body")
		return
	}
	user.Nickname = nickname

	err = h.repo.EditUser(user)
	if err != nil {
		buildmode.Log.Println("Error in UpdateUser-EditUser: ", err)
		switch err.Error() {
		case structs.ErrorDuplicateKey:
			resp.
				SetStatus(409).
				SetContent(struct{
					Message string `json:"message"`
				}{Message:"Email is used " + user.Email})
			return
		default:
			resp.
				SetStatus(404).
				SetContent(struct{
					Message string `json:"message"`
				}{Message:"Can-t find user with nickname " + nickname})
			return
		}
	}
	user, err = h.repo.GetUser("", nickname)
	if err != nil {
		buildmode.Log.Println("Error in UpdateUser-GetUser: ", err)
		resp.
			SetStatus(404).
			SetContent(struct{
				Message string `json:"message"`
			}{Message:"Can-t find user with nickname " + nickname})
		return
	}
	resp.SetStatus(200).SetError(user)
	return
}