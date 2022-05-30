package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"offlinemapexp/render"

	"github.com/gorilla/mux"
	lru "github.com/hashicorp/golang-lru"
)

var queue render.RenderQueue
var cache *lru.Cache

const prerenderedZoom = 10

func tileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	x, err := strconv.Atoi(vars["x"])
	if err != nil {
		return
	}
	y, err := strconv.Atoi(vars["y"])
	if err != nil {
		return
	}
	z, err := strconv.Atoi(vars["z"])
	if err != nil {
		return
	}
	//Is tile in the cache?
	if cache.Contains(render.Coords{X: x, Y: y, Z: z}) {
		image, exists := cache.Get(render.Coords{X: x, Y: y, Z: z})
		if exists {
			w.Write(image.([]byte))
			return
		}
	}
	//Is it prerendered?
	if z >= prerenderedZoom {
		filename := "prerendered/" + strconv.Itoa(x) + "_" + strconv.Itoa(y) + "_" + strconv.Itoa(z) + ".png"
		image, err := ioutil.ReadFile(filename)
		if err == nil {
			w.Write(image)
			go cache.Add(render.Coords{X: x, Y: y, Z: z}, image)
			return
		}
	}
	//Render
	tileRender := queue.GetTileRender()
	defer queue.PutTileRender(tileRender)
	image, err := tileRender.Render(x, y, z)
	if len(image) < 200 {
		fmt.Println("Bad render for ", render.Coords{X: x, Y: y, Z: z}, len(image))
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Write(image)
	go cache.Add(render.Coords{X: x, Y: y, Z: z}, image)
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Mainpage")
	http.ServeFile(w, r, "js/main.html")
}

func main() {
	var err error
	cache, err = lru.New(1024)
	if err != nil {
		log.Fatal(err)
	}
	queue.InitQueue(10, "style.xml")
	router := mux.NewRouter()
	router.HandleFunc("/", mainPageHandler)
	router.HandleFunc("/{z:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}.png", tileHandler).Methods("GET")

	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))

	srv := &http.Server{
		Handler: http.TimeoutHandler(router, time.Second*30, ""),
		Addr:    "127.0.0.1:8080",
	}
	log.Fatal(srv.ListenAndServeTLS("server.crt", "server.key"))
}
