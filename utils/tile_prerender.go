package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"offlinemapexp/render"
	"os"
	"strconv"
	"sync"
)

var queue render.RenderQueue

const poolSize = 10

func renderPool(ch <-chan render.Coords, wg *sync.WaitGroup) {
	for coords := range ch {
		tileRender := queue.GetTileRender()
		tileRender.RenderToFile(coords.X, coords.Y, coords.Z)
		queue.PutTileRender(tileRender)
	}
	wg.Done()
}

func main() {
	var zoom = flag.Int("z", 0, "zoom level")
	var xCenter = flag.Int("x", 0, "center X position")
	var yCenter = flag.Int("y", 0, "center Y position")
	var radius = flag.Int("r", 1, "radius for prerender")
	var xmlFile = flag.String("f", "", "xml style file for mapnik")
	flag.Parse()
	if *zoom < 0 || *xmlFile == "" {
		fmt.Println("Enter zoom level and xml style file for mapnik")
		return
	}
	if *xCenter-*radius < 0 || *yCenter-*radius < 0 {
		fmt.Println("Enter valid X, Y and Radius values")
		return
	}
	wg := sync.WaitGroup{}
	queue.InitQueue(poolSize, "style.xml")
	coordSender := make(chan render.Coords)
	for i := 0; i < poolSize; i++ {
		wg.Add(1)
		go renderPool(coordSender, &wg)
	}
	err := os.MkdirAll("prerendered/"+strconv.Itoa(*zoom), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	for x := *xCenter - *radius; x < int(math.Pow(2, float64(*zoom))) && x < *xCenter+*radius; x++ {
		for y := *yCenter - *radius; y < int(math.Pow(2, float64(*zoom))) && y < *yCenter+*radius; y++ {
			coordSender <- render.Coords{X: x, Y: y, Z: *zoom}
		}
	}
	close(coordSender)
	wg.Wait()
	fmt.Println("Done")
}
