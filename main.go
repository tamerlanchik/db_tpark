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
	//type S struct {
	//	A string `json:"a,omitempty"`
	//	B int `json:"b,omitempty"`
	//}
	//
	//w := S{A:""}
	//
	//res, _ := json.Marshal(w)
	//fmt.Println(string(res))



	mainRouter := mux.NewRouter().PathPrefix("/api").Subrouter()
	InflateRouter(mainRouter)

	err := http.ListenAndServe(":"+port, mainRouter)
	fmt.Println("Stop server: ", err)
}

func InflateRouter(r *mux.Router) error {
	r.Use(AddContentTypeMiddleware())

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

func AddContentTypeMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.URL)
			next.ServeHTTP(w, r)
			fmt.Println(w.Header())
		})

	}
}


//w.Header().Set("Content-Type", "application/json")