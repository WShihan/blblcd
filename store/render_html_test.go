package store

import (
	"testing"
)

func TestRenderHtm(t *testing.T) {
	_, err := RenderHTML("评论分布", "", "")
	if err != nil {
		panic(err)
	}
}
