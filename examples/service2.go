package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	botzilla "github.com/Pois-Noir/Botzilla"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	mecca500, err := botzilla.JointNamespace("mecca500", "meow", logger, false)
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	output, err := botzilla.Call[string, int](mecca500, ctx, "s1", "amir")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(output)

}
