package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/Pois-Noir/Botzilla"
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

	s1Handler := func(input string) (output int, err error) {
		return 21, nil
	}

	_, err = botzilla.NewService(mecca500, "s1", s1Handler)

	if err != nil {
		log.Fatal(err)
	}

	// Note that the call can be in separate executable or even machine. as long as two machines are connected in the same
	// local network and have zero conf library discovery service finds the other end point

	select {}

}
