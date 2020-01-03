package main

import (
	"db_tpark/buildmode"
	"db_tpark/handler"
	"db_tpark/repository"
	"runtime"
	"time"

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
	buildmode.LogTag = "no"
	buildmode.Log.Println("Start server ", runtime.GOMAXPROCS(0))
	//runtime.GOMAXPROCS(6)
	mainRouter := mux.NewRouter().PathPrefix("/api").Subrouter()
	if err:= InflateRouter(mainRouter); err !=nil {
		buildmode.Log.Println("Error inflating router:", err)
		if buildmode.BuildTag=="debug" {
			return
		}
	}
	//buildmode.Log.Println("Router inflated")

	//mainRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	w.Write([]byte("This is a catch-all route"))
	//	//buildmode.Log.Println("FF", r.URL)
	//})
	err := http.ListenAndServe(":"+port, mainRouter)
	buildmode.Log.Println("Stop server: ", err)
}

func InflateRouter(r *mux.Router) error {

	if buildmode.BuildTag=="debug" {
		r.Use(AddContentTypeMiddleware())
	}



	repo := repository.NewPostgresRepo()
	err := repo.Init(dbuser, dbpass, dbhost, dbport, dbname);
	if err != nil {
		//buildmode.Log.Println("Error during init postgres")
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

	r.HandleFunc("/metrics", func(writer http.ResponseWriter, request *http.Request) {
		handler.PrintMetrics()
	})


	return nil
}

func AddContentTypeMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buildmode.Log.Println("Path: ", r.URL)
			start := time.Now()
			next.ServeHTTP(w, r)
			buildmode.Log.Println(r.URL, " ", time.Since(start))
		})

	}
}