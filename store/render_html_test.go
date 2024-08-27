package store

import (
	"testing"
)

func TestRenderHtm(t *testing.T) {
	_, err := RenderHTML("", "")
	if err != nil {
		panic(err)
	}
}
