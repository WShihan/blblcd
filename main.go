package main

import "blblcd/cli"

var (
	Version   string
	BuildTime string
	Commit    string
	Author    = "Wangshihan"
)

func main() {
	cli.Execute(&cli.Injection{
		Version:   Version,
		BuildTime: BuildTime,
		Commit:    Commit,
		Author:    Author,
	})
}
