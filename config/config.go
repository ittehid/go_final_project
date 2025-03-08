package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func GetDBFilePath() string {
	if envDBFile := os.Getenv("TODO_DBFILE"); envDBFile != "" {
		return envDBFile
	}

	_, filename, _, _ := runtime.Caller(0)
	projectDir := filepath.Dir(filepath.Dir(filename))

	dataDir := filepath.Join(projectDir, "data")
	os.MkdirAll(dataDir, os.ModePerm)

	return filepath.Join(dataDir, "scheduler.db")
}
