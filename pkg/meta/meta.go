package meta

import (
	"encoding/json"
	"os"
	"time"
)

type Meta struct {
	ExpoStart   string `json:"expoStart"`
	GeneratedAt string `json:"generatedAt"`
}

// Write encodes m as JSON and writes it to path, creating or truncating the file.
func Write(path, expoStart string, generatedAt time.Time) error {
	m := Meta{
		ExpoStart:   expoStart,
		GeneratedAt: generatedAt.UTC().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
