package main

import (
	"fmt"
	"lib/log"
	"lib/tool"
	"time"
)

func main() {
	if err := tool.InitModule("./conf/dev/", []string{"base", "mysql", "redis"}); err != nil {
		log.Fatal(fmt.Sprintf("%s", err))
	}
	defer tool.Destroy()

	tool.Log.TagInfo(tool.NewTrace(), tool.DLTagUndefind, map[string]interface{}{
		"message": "todo something",
	})
	time.Sleep(time.Second)
}
