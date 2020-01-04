package main

import (
	"db_tpark/buildmode"
	"db_tpark/handler"
	"db_tpark/repository"
	"fmt"
	_ "net/http/pprof"
	"runtime"
	"time"

	//"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

//	http://localhost:5000/api/thread/37/create

const (
	dbuser = "docker"
	dbpass = "docker"
	dbhost = "0.0.0.0"
	dbport = "5432"
	dbname = "dbtpark"
	port = "5000"
)

//	http://localhost:5000/api/thread/1857/create

func garbageCOllectionTask(timeout int) {
	ticker := time.NewTicker(time.Millisecond*time.Duration(timeout))
	go func() {
		for _ = range ticker.C {
			runtime.GC()
		}
	}()
}

func main() {
	buildmode.LogTag = "log"
	fmt.Println("Start server ", runtime.GOMAXPROCS(0))
	start := mux.NewRouter()
	mainRouter := start.PathPrefix("/api").Subrouter()
	if err:= InflateRouter(mainRouter); err !=nil {
		fmt.Println("Error inflating router:", err)
		if buildmode.BuildTag=="debug" {
			return
		}
	}
	//go http.ListenAndServe("localhost:5001", nil)

	//start.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	err := http.ListenAndServe(":"+port, mainRouter)
	fmt.Println("Stop server: ", err)
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

var ttt int64

func AddContentTypeMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buildmode.Log.Println("Path: ", r.URL)
			//ttt++
			//start := time.Now()
			//var m runtime.MemStats
			//runtime.ReadMemStats(&m)
			//// For info on each, see: https://golang.org/pkg/runtime/#MemStats
			//before := int64(m.Alloc)
			next.ServeHTTP(w, r)
			//runtime.GC()
			//runtime.ReadMemStats(&m)
			//buildmode.Log.Println(r.URL, " ", time.Since(start))
			//if (int64(m.Alloc)-before)/1024 > 100{
			//	fmt.Println(	"Memory added: ", (int64(m.Alloc)-before)/1024)
			//	if ttt%1000==0{
			//		runtime.GC()
			//		fmt.Println("Clear")
			//	}
			//}
			//runtime.GC()
		})

	}
}