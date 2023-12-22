package main

import (
	_ "meting-api/music/netease"
	_ "meting-api/music/tencent"

	"meting-api/internal/cmd"
)

func main() {
	cmd.Run()
}
