package main

import (
	"db_tpark/handler"
	"db_tpark/repository"
	"fmt"
	"runtime"

	//"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	dbuser = "docker"
	dbpass = "docker"
	dbhost = "0.0.0.0"
	dbport = "5432"
	dbname = "dbtpark"
	port = "5000"
)

func main() {
	fmt.Println("Start server ", runtime.GOMAXPROCS(0))
	mainRouter := mux.NewRouter().PathPrefix("/api").Subrouter()
	if err:= InflateRouter(mainRouter); err !=nil {
		//fmt.Println("Error inflating router:", err)
		//return
	}
	//fmt.Println("Router inflated")

	mainRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a catch-all route"))
		//fmt.Println("FF", r.URL)
	})
	loggedRouter := AddContentTypeMiddleware()(mainRouter)
	http.Handle("/", loggedRouter)
	//fmt.Println("Prestart server")
	err := http.ListenAndServe(":"+port, loggedRouter)
	fmt.Println("Stop server: ", err)
}

func InflateRouter(r *mux.Router) error {

	//r.Use(AddContentTypeMiddleware())



	repo := repository.NewPostgresRepo()
	err := repo.Init(dbuser, dbpass, dbhost, dbport, dbname);
	if err != nil {
		//fmt.Println("Error during init postgres")
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
			fmt.Println("Path: ", r.URL)
			next.ServeHTTP(w, r)
		})

	}
}