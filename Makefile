.PHONY: test build

export CGO_LDFLAGS = $(shell mapnik-config --libs)
export CGO_CXXFLAGS = $(shell mapnik-config --cxxflags --includes --dep-includes | tr '\n' ' ')

MAPNIK_LDFLAGS=-X github.com/KanesRF/go-mapnik/v3.fontPath=$(shell mapnik-config --fonts) \
	-X github.com/KanesRF/go-mapnik/v3.pluginPath=$(shell mapnik-config --input-plugins)

backend:
	go build -o cmd/server -ldflags "$(MAPNIK_LDFLAGS)" server/server.go

prerender:
	go build -o cmd/prerender -ldflags "$(MAPNIK_LDFLAGS)" utils/tile_prerender.go

all: backend prerender