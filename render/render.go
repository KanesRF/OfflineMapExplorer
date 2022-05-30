package render

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"

	mapnik "github.com/KanesRF/go-mapnik/v3"
)

const (
	boundX0 = -20037508.3428
	boundX1 = 20037508.3428
	boundY0 = -20037508.3428
	boundY1 = 20037508.3428
)

type Coords struct {
	X, Y, Z int
}

func makeCoords(x, y float64, z int) (p0x, p0y, p1x, p1y float64) {
	zoom := 1 << z
	p0x = boundX0 + (boundX1-boundX0)/float64(zoom)*x
	p0y = boundY1 - (boundY1-boundY0)/float64(zoom)*(y+1)
	p1x = boundX0 + (boundX1-boundX0)/float64(zoom)*(x+1)
	p1y = boundY1 - (boundY1-boundY0)/float64(zoom)*(y)
	return p0x, p0y, p1x, p1y
}

type TileRender struct {
	m    mapnik.Map
	opts mapnik.RenderOpts
}

type RenderQueue struct {
	queue       []*TileRender
	isAvailable sync.Cond
}

func (renderQueue *RenderQueue) InitQueue(size int, xmlConfig string) {
	renderQueue.queue = make([]*TileRender, size, size)
	for i := range renderQueue.queue {
		renderQueue.queue[i] = new(TileRender)
		if err := renderQueue.queue[i].InitRender(xmlConfig); err != nil {
			log.Fatal(err)
		}
	}
	renderQueue.isAvailable.L = &sync.Mutex{}
	fmt.Println("Inited RenderQueue")
	fmt.Println("The number of CPU Cores:", runtime.NumCPU())
}

func (renderQueue *RenderQueue) PutTileRender(render *TileRender) {
	renderQueue.isAvailable.L.Lock()
	defer renderQueue.isAvailable.L.Unlock()
	renderQueue.queue = append(renderQueue.queue, render)
	renderQueue.isAvailable.Signal()
}

func (renderQueue *RenderQueue) GetTileRender() *TileRender {
	renderQueue.isAvailable.L.Lock()
	defer renderQueue.isAvailable.L.Unlock()
	for len(renderQueue.queue) == 0 {
		renderQueue.isAvailable.Wait()
	}
	curTileRender := renderQueue.queue[len(renderQueue.queue)-1]
	renderQueue.queue = renderQueue.queue[:len(renderQueue.queue)-1]
	return curTileRender
}

func (render *TileRender) InitRender(stylePath string) error {
	render.m = *mapnik.New()
	if err := render.m.Load(stylePath); err != nil {
		return err
	}
	render.m.Resize(256, 256)
	render.opts = mapnik.RenderOpts{Format: "png32"}
	return nil
}

func (render *TileRender) RenderToFile(x, y, z int) {
	x0, y0, x1, y1 := makeCoords(float64(x), float64(y), z)
	render.m.ZoomTo(x0, y0, x1, y1)
	if err := render.m.RenderToFile(render.opts, "prerendered/"+strconv.Itoa(x)+"_"+strconv.Itoa(y)+"_"+strconv.Itoa(z)+".png"); err != nil {
		log.Fatal(err)
	}
}

func (render *TileRender) Render(x, y, z int) ([]byte, error) {
	x0, y0, x1, y1 := makeCoords(float64(x), float64(y), z)
	render.m.ZoomTo(x0, y0, x1, y1)
	return render.m.Render(render.opts)
}
