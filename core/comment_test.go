package core

import (
	"fmt"
	"log"
	"os"
	"testing"
)

var (
	Cookie string
)

func init() {
	data, err := os.ReadFile("../cookie.text")
	if err != nil {
		log.Fatal(err)
	}
	Cookie = string(data)
}

func TestFetchCmt(t *testing.T) {
	oid := Bvid2Avid("BV1XU1eYTEW4")
	t.Log(FetchComment(fmt.Sprintf("%d", oid), 0, 2, Cookie))
}
func TestFetchSubCmt(t *testing.T) {
	oid := Bvid2Avid("BV1XU1eYTEW4")
	t.Log(FetchSubComment(fmt.Sprintf("%d", oid), 243795113873, 13, Cookie))
}
