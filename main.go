package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Yokanater/spotr/internal/app"
	"github.com/Yokanater/spotr/internal/paths"
	"github.com/Yokanater/spotr/store"
)

var version = "dev"

func main() {
	dataDir := flag.String("data-dir", "", "directory used to store spotr data")
	showVersion := flag.Bool("version", false, "print the spotr version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("spotr %s\n", version)
		return
	}

	dbPath, err := paths.DatabasePath(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "spotr: data directory: %v\n", err)
		os.Exit(1)
	}

	st, err := store.NewSQLite(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "spotr: open db: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	if err := app.Run(st); err != nil {
		fmt.Fprintf(os.Stderr, "spotr: %v\n", err)
		os.Exit(1)
	}
}
