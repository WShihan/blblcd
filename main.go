package main

import (
	"blblcd/cli"
	"blblcd/model"
)

var (
	Version   string
	BuildTime string
	Commit    string
	Author    = "Wangshihan"
)

func main() {
	cli.Execute(&model.Injection{
		Version:   Version,
		BuildTime: BuildTime,
		Commit:    Commit,
		Author:    Author,
	})
}
