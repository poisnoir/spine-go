package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/poisnoir/spine-go"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ns, _ := spine.JointNamespace("example", "meow", logger, false)

	handle1 := func(temp uint32) {
		fmt.Printf("sub 1: %d\n", temp)
	}

	//handle2 := func(temp uint32) {
	//	fmt.Printf("sub 2 : %d\n", temp)
	//}

	_, _ = spine.NewSubscriber(ns, "temperature", handle1)
	//_, _ = spine.NewSubscriber(ns, "temperature", handle2)

	select {}
}
