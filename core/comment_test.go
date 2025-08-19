package core

import (
	"log"
	"os"
	"strconv"
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
	t.Log(FetchComment(strconv.FormatInt(oid, 10), 2, Cookie, ""))
}
func TestFetchSubCmt(t *testing.T) {
	oid := Bvid2Avid("BV1XU1eYTEW4")
	t.Log(FetchSubComment(strconv.FormatInt(oid, 10), 243795113873, 13, Cookie))
}
func TestFetchSubCmtCount(t *testing.T) {
	oid := Bvid2Avid("BV1XU1eYTEW4")
	t.Log(FetchCount(strconv.FormatInt(oid, 10)))
}
