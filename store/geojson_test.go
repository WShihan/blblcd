package store

import (
	"testing"
)

func TestGeojsonWrite(t *testing.T) {
	statMap := map[string]int{
		"广东": 100,
		"北京": 10,
	}
	WriteGeoJSON(statMap, "123", "output")
}
