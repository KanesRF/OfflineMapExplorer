package main

import (
	"flag"
	"fmt"
	"log"
	"offlinemapexp/render"
	"os"
	"strconv"
	"sync"
)

const poolSize = 10

func renderPool(ch <-chan render.Coords, wg *sync.WaitGroup, queue *render.Queue) {
	for coords := range ch {
		tileRender := queue.GetTileRender()
		if err := tileRender.RenderToFile(coords.X, coords.Y, coords.Z); err != nil {
			fmt.Println(err)
		}
		queue.PutTileRender(tileRender)
	}
	wg.Done()
}

func main() {
	var queue render.Queue
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
	queue.InitQueue(poolSize, *xmlFile)
	coordSender := make(chan render.Coords)
	for i := 0; i < poolSize; i++ {
		wg.Add(1)
		go renderPool(coordSender, &wg, &queue)
	}
	err := os.MkdirAll("prerendered/"+strconv.Itoa(*zoom), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	var startX, startY int
	if startX = *xCenter - *radius; startX < 0 {
		startX = 0
	}
	if startY = *yCenter - *radius; startY < 0 {
		startY = 0
	}
	for x := startX; x < 1<<*zoom && x < *xCenter+*radius; x++ {
		for y := startY; y < 1<<*zoom && y < *yCenter+*radius; y++ {
			coordSender <- render.Coords{X: x, Y: y, Z: *zoom}
		}
	}
	close(coordSender)
	wg.Wait()
	fmt.Println("Done")
}
