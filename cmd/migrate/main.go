package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

	"clean-template/internal/config"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	var (
		dir    = flag.String("dir", "migrations", "migrations directory")
		command string
	)
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: migrate <command>")
		os.Exit(1)
	}
	command = args[0]

	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		fatal(err)
	}

	db, err := sql.Open("postgres", cfg.Postgres.DSN())
	if err != nil {
		fatal(err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		fatal(err)
	}

	ctx := context.Background()
	switch command {
	case "up":
		err = goose.UpContext(ctx, db, *dir)
	case "down":
		err = goose.DownContext(ctx, db, *dir)
	case "status":
		err = goose.StatusContext(ctx, db, *dir)
	case "reset":
		err = goose.ResetContext(ctx, db, *dir)
	default:
		fatal(fmt.Errorf("unknown command: %s", command))
	}
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
	os.Exit(1)
}
