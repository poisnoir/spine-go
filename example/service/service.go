package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/poisnoir/spine-go"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ns, _ := spine.JointNamespace("example", "meow", logger)

	lenFunc := func(input string) (uint32, error) {
		return uint32(len(input)), nil
	}

	printFunc := func(input string) (string, error) {
		fmt.Println(input)
		return "printed " + input, nil
	}

	_, _ = spine.NewService(ns, "string_length", lenFunc)
	_, _ = spine.NewService(ns, "print", printFunc)

	select {}

}
