package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	mapnik "github.com/omniscale/go-mapnik"
)

var connectString = "user=map_app dbname=map_db password=12345 host=localhost sslmode=disable"
var DbConn *sql.DB

const (
	Metafile = 8
	BoundX0  = -20037508.3428
	BoundX1  = 20037508.3428
	BoundY0  = -20037508.3428
	BoundY1  = 20037508.3428
)

func InitDB() {
	var err error
	DbConn, err = sql.Open("postgres", connectString)
	if err != nil {
		panic(err)
	}
}

func CloseDB() {
	DbConn.Close()
}

func MakeCoords(x, y float64, z int) (p0x, p0y, p1x, p1y float64) {
	zoom := 1 << z
	p0x = BoundX0 + (BoundX1-BoundX0)/float64(zoom)*x
	p0y = BoundY1 - (BoundY1-BoundY0)/float64(zoom)*(y+1)
	p1x = BoundX0 + (BoundX1-BoundX0)/float64(zoom)*(x+1)
	p1y = BoundY1 - (BoundY1-BoundY0)/float64(zoom)*(y)
	fmt.Println(x, y, z, " To ", p0x, p0y, p1x, p1y)
	return p0x, p0y, p1x, p1y
}

func DoAll(x, y, z float64) {
	x0, y0, x1, y1 := MakeCoords(x, y, int(z))
	m := mapnik.New()
	if err := m.Load("mapnik_cfg.xml"); err != nil {
		log.Fatal(err)
	}
	m.Resize(256, 256)
	m.ZoomTo(x0, y0, x1, y1)
	opts := mapnik.RenderOpts{Format: "png32"}
	if err := m.RenderToFile(opts, "example.png"); err != nil {
		log.Fatal(err)
	}

}

func main() {
	InitDB()
	DoAll(3, 1, 2)
}
