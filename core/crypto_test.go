package core

import (
	"testing"
)

func TestAvid2Bvid(t *testing.T) {
	res := Avid2Bvid(113976755094913)
	t.Log(res)
}

func TestBvid2Avid(t *testing.T) {
	res := Bvid2Avid("BV1e7NRemEwv")
	t.Log(res)
}
