package main

import (
	"db_tpark/handler"
	"db_tpark/repository"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	dbuser = "postgres"
	dbpass = "postgres"
	dbhost = "0.0.0.0"
	dbport = "5432"
	dbname = "dbtpark"
	port = "5000"
)

func main() {
	mainRouter := mux.NewRouter()
	InflateRouter(mainRouter)

	err := http.ListenAndServe(":"+port, mainRouter)
	fmt.Println("Stop server: ", err)
}

func InflateRouter(r *mux.Router) error {
	repo := repository.NewPostgresRepo()
	err := repo.Init(dbuser, dbpass, dbhost, dbport, dbname);
	if err != nil {
		return err
	}

	forumRouter := r.PathPrefix("/forum").Subrouter()
	handler.NewForumHandler(repo).InflateRouter(forumRouter)

	postRouter := r.PathPrefix("/post").Subrouter()
	handler.NewPostHandler(repo).InflateRouter(postRouter)

	serviceRouter := r.PathPrefix("/service").Subrouter()
	handler.NewServiceHandler(repo).InflateRouter(serviceRouter)

	threadRouter := r.PathPrefix("/thread").Subrouter()
	handler.NewThreadHandler(repo).InflateRouter(threadRouter)

	userRouter := r.PathPrefix("/user").Subrouter()
	handler.NewUserHandler(repo).InflateRouter(userRouter)

	return nil
}