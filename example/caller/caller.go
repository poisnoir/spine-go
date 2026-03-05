package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/poisnoir/spine-go"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ns, _ := spine.JointNamespace("example", "meow", logger, false)

	ctx, _ := context.WithCancel(context.Background())

	// will error if types are mismatched
	c, _ := spine.NewServiceCaller[string, uint32](ns, "string_length")

	// will error if it can't get result before context cancels
	// Blocks until result is received or context is canceled
	result, _ := c.Call("hello world", ctx)
	fmt.Println(result)

	select {}
}
