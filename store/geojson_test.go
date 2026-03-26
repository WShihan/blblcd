package store

import (
	"blblcd/model"
	"testing"
)

func TestGeojsonWrite(t *testing.T) {
	statMap := map[string]model.Stat{}
	WriteGeoJSON(statMap, "123", "output")
}
