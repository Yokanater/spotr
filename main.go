package main

import (
	"fmt"
	"os"
	"ruffnut/internal/app"
	"ruffnut/store"
)

func main() {
	st, err := store.NewSQLite("ruffnut.db")
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
