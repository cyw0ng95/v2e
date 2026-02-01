package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

func main() {
	dbPath := flag.String("db", "./notes.db", "Path to SQLite DB file")
	dir := flag.String("dir", "tool/migrations", "Migrations directory")
	flag.Parse()

	files := []string{}
	_ = filepath.WalkDir(*dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(p) == ".sql" {
			files = append(files, p)
		}
		return nil
	})

	sort.Strings(files)

	for _, f := range files {
		fmt.Printf("Applying migration: %s\n", f)
		cmd := exec.Command("sqlite3", *dbPath, ".read "+f)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Migrations applied successfully")
}
