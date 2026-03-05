package main

import (
	"log/slog"
	"os"

	"github.com/poisnoir/spine-go"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ns, _ := spine.JointNamespace("example", "meow", logger, false)

	handler := func(input string) (uint32, error) {
		return uint32(len(input)), nil
	}

	_, _ = spine.NewService(ns, "string_length", handler)

	select {}

}
