package render

import (
	"log"
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
	m    *mapnik.Map
	opts *mapnik.RenderOpts
}

type Queue struct {
	queue       []*TileRender
	isAvailable sync.Cond
}

func (renderQueue *Queue) InitQueue(size int, xmlConfig string) error {
	var err error = nil
	renderQueue.queue = make([]*TileRender, size, size)
	errChan := make(chan error, len(renderQueue.queue))
	defer close(errChan)
	for i := range renderQueue.queue {
		renderQueue.queue[i] = new(TileRender)
		go func(tile *TileRender) {
			err := tile.InitRender(xmlConfig)
			errChan <- err
		}(renderQueue.queue[i])
	}
	for i := 0; i < len(renderQueue.queue); i++ {
		tmpErr := <-errChan
		if tmpErr != nil {
			err = tmpErr
		}
	}
	renderQueue.isAvailable.L = &sync.Mutex{}
	log.Printf("Inited Queue")
	return err
}

func (renderQueue *Queue) PutTileRender(render *TileRender) {
	renderQueue.isAvailable.L.Lock()
	defer renderQueue.isAvailable.L.Unlock()
	renderQueue.queue = append(renderQueue.queue, render)
	renderQueue.isAvailable.Signal()
}

func (renderQueue *Queue) GetTileRender() *TileRender {
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
	render.m = mapnik.New()
	if err := render.m.Load(stylePath); err != nil {
		return err
	}
	render.m.Resize(256, 256)
	render.opts = &mapnik.RenderOpts{Format: "png32"}
	return nil
}

func (render *TileRender) RenderToFile(x, y, z int) error {
	x0, y0, x1, y1 := makeCoords(float64(x), float64(y), z)
	render.m.ZoomTo(x0, y0, x1, y1)
	if err := render.m.RenderToFile(*render.opts, "prerendered/"+strconv.Itoa(z)+"/"+strconv.Itoa(x)+"_"+strconv.Itoa(y)+".png"); err != nil {
		return err
	}
	return nil
}

func (render *TileRender) Render(x, y, z int) ([]byte, error) {
	x0, y0, x1, y1 := makeCoords(float64(x), float64(y), z)
	render.m.ZoomTo(x0, y0, x1, y1)
	return render.m.Render(*render.opts)
}
