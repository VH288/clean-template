package main

import (
	"log"

	"clean-template/internal/app"
)

func main() {
	if err := app.New().Run(); err != nil {
		log.Fatal(err)
	}
}
