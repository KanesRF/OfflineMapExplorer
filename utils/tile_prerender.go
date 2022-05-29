package main

import (
	"flag"
	"fmt"
	"math"
	"offlinemapexp/render"
)

var queue render.RenderQueue

func renderPool(ch <-chan render.Coords) {
	for coords := range ch {
		tileRender := queue.GetTileRender()
		tileRender.RenderToFile(coords.X, coords.Y, coords.Z)
		queue.PutTileRender(tileRender)
	}
}

func main() {
	var zoom = flag.Int("z", 0, "zoom level")
	var xCenter = flag.Int("x", 0, "center X position")
	var yCenter = flag.Int("y", 0, "center Y position")
	var radius = flag.Int("r", 40, "radius for prerender")
	var xmlFile = flag.String("f", "", "xml style file for mapnik")
	flag.Parse()
	if *zoom < 0 || *xmlFile == "" {
		fmt.Println("Enter zoom level and xml style file for mapnik")
		return
	}

	queue.InitQueue(20, "style.xml")
	coordSender := make(chan render.Coords)
	for i := 0; i < 10; i++ {
		go renderPool(coordSender)
	}
	for x := *xCenter - *radius; x < int(math.Pow(2, float64(*zoom))) && x < *xCenter+*radius; x++ {
		for y := *yCenter - *radius; y < int(math.Pow(2, float64(*zoom))) && y < *yCenter+*radius; y++ {
			coordSender <- render.Coords{X: x, Y: y, Z: *zoom}
		}
	}
	close(coordSender)
}
