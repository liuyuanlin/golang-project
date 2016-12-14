package main

import (
	"flag"
	"golang-project/dpsg/csvcfg"
	"golang-project/dpsg/logger"
)

type BuildingCfg struct {
	Name              string
	TID               string
	InfoTID           string
	BuildingClass     string
	ResName           string
	BuildSize         uint32
	BuildTimeD        uint32 `csv:",d"`
	BuildTimeH        uint32 `csv:",d"`
	BuildTimeM        uint32 `csv:",d"`
	BuildResource     string
	BuildCost         uint32
	Icon              string
	Hitpoints         uint32
	ProducesResource  string
	ResourcePerHour   uint32
	ResourceMax       uint32
	ResourceIconLimit uint32
	BoostCost         uint32
	MaxStoredGold     uint32
	MaxStoredFood     uint32
}

var (
	csvFile = flag.String("c", "in.csv", "csv input file")
)

func main() {

	flag.Parse()

	var aa map[string]*[]BuildingCfg

	csvcfg.LoadCSVConfig(*csvFile, &aa)

	for k, v := range aa {
		logger.Info("key %v", k)
		for ii, vv := range *v {
			logger.Info("%i %v", ii, vv)
		}
	}
}
