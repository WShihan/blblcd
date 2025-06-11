package utils

import (
	"blblcd/model"
	"testing"
)

func TestFileOrPathExists(t *testing.T) {
	exits := FileOrPathExists("./tool.go")
	if exits != true {
		t.Error("文件不存在")
	}
	notExits := FileOrPathExists("./tool2.go")
	if notExits != false {
		t.Error("文件存在")
	}
	t.Log(exits)
}

func TestReadTextFile(t *testing.T) {
	res, err := ReadTextFile("../cookie.text")
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
}
func TestEncodePaginationOffset(t *testing.T) {
	res := EncodePaginationOffset(model.PaginationOffset{
		Type:      1,
		Direction: 1,
		Data: model.PaginationOffsetData{
			Pn: 1,
		}})
	t.Log(res)
}

func TestDecodePaginationOffset(t *testing.T) {
	res, errr := DecodePaginationOffset("{\"type\":1,\"direction\":1,\"data\":{\"pn\":2}}")
	if errr != nil {
		t.Error(errr)
	}
	t.Logf("%v", res)
	t.Log(res)
}

func TestProgressBar(t *testing.T) {
	ProgressBar(500, 1000)
	t.Log("")
}

func TestPrintLogo(t *testing.T) {
	PrintLogo()
	t.Log("")
}
