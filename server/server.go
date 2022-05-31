package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"offlinemapexp/render"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	lru "github.com/hashicorp/golang-lru"
)

var (
	queue           render.Queue
	cache           *lru.Cache
	prerenderedZoom int = 7
	poolSize        int = 4
	maxZoom         int = 10
	xmlPath         string
)

func validateCoords(x, y, z int) error {
	if z < 0 || z > maxZoom {
		return errors.New("Invalid Z")
	}
	if x < 0 || x >= int(math.Pow(2, float64(z))) {
		return errors.New("Invalid X")
	}
	if y < 0 || y >= int(math.Pow(2, float64(z))) {
		return errors.New("Invalid Y")
	}
	return nil
}

func tileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	x, err := strconv.Atoi(vars["x"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	y, err := strconv.Atoi(vars["y"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	z, err := strconv.Atoi(vars["z"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := validateCoords(x, y, z); err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
		filename := "prerendered/" + strconv.Itoa(z) + "/" + strconv.Itoa(x) + "_" + strconv.Itoa(y) + ".png"
		image, err := ioutil.ReadFile(filename)
		if err == nil {
			w.Write(image)
			cache.Add(render.Coords{X: x, Y: y, Z: z}, image)
			return
		}
	}
	//Render
	tileRender := queue.GetTileRender()
	defer queue.PutTileRender(tileRender)
	image, err := tileRender.Render(x, y, z)
	if err != nil {
		w.Write([]byte("Error rendering"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(image)
	cache.Add(render.Coords{X: x, Y: y, Z: z}, image)
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "js/main.html")
}

func main() {

	flag.IntVar(&prerenderedZoom, "z", prerenderedZoom, "min zoom of prerendered tiles")
	flag.IntVar(&poolSize, "pool", poolSize, "pool of Map objects for rendering")
	flag.IntVar(&maxZoom, "max_zoom", maxZoom, "max zoom")
	flag.StringVar(&xmlPath, "f", "", "xml style file for mapnik")
	flag.Parse()
	if xmlPath == "" {
		fmt.Println("Enter path for XML Mapnik file")
		return
	}
	var err error
	cache, err = lru.New(1024)
	if err != nil {
		log.Fatal(err)
	}
	queue.InitQueue(poolSize, xmlPath)
	router := mux.NewRouter()
	router.HandleFunc("/", mainPageHandler)
	router.HandleFunc("/{z:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}.png", tileHandler).Methods("GET")
	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))

	srv := &http.Server{
		Handler:     http.TimeoutHandler(router, time.Second*50, ""),
		Addr:        "127.0.0.1:8080",
		ReadTimeout: time.Second * 20,
	}
	log.Fatal(srv.ListenAndServeTLS("server.crt", "server.key"))
}
