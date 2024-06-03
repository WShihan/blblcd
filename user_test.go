package main

import (
	"blblcd/core"
	"fmt"
	"log/slog"
	"testing"
)

func CoreTest(t *testing.T) {
	data, err := core.FetchVideoList(1556651916, 1, "", "cookie")
	if err != nil {
		slog.Error(err.Error())
	}
	fmt.Printf("%v", data)
}
