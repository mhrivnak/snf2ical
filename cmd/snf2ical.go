package main

import (
	"encoding/json"
	"fmt"
	"io"

	"os"
	"path/filepath"
	"time"

	"github.com/mhrivnak/snf2ical/pkg/parse"
	"github.com/mhrivnak/snf2ical/pkg/status"

	"github.com/spf13/cobra"
)

var OutDir string
var Year int
var ScheduleURL string
var JSONFile string

func main() {
	rootCmd := &cobra.Command{
		Use:   "snf2ical",
		Short: "generate ical files for Sun n Fun",
		Long:  "snf2ical creates or overwrites several ical files for Sun n Fun",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var rows []parse.Row
			var err error

			// Validate mutually exclusive flags
			if ScheduleURL != "" && JSONFile != "" {
				fmt.Println("error: cannot specify both --url and --json")
				os.Exit(1)
			}

			// Determine which input source to use based on flags
			if ScheduleURL != "" {
				// Fetch HTML from URL
				fmt.Printf("Fetching schedule from %s...\n", ScheduleURL)
				rows, err = parse.FetchScheduleHTML(ScheduleURL)
				if err != nil {
					fmt.Printf("failed to fetch schedule from URL: %v\n", err)
					os.Exit(1)
				}
			} else if JSONFile != "" {
				// Parse JSON from file
				f, err := os.Open(JSONFile)
				if err != nil {
					fmt.Printf("failed to open JSON file: %v\n", err)
					os.Exit(1)
				}
				defer f.Close()

				data, err := io.ReadAll(f)
				if err != nil {
					fmt.Printf("failed to read JSON file: %v\n", err)
					os.Exit(1)
				}

				rows = make([]parse.Row, 0)
				if err = json.Unmarshal(data, &rows); err != nil {
					fmt.Printf("failed to unmarshal JSON file: %v\n", err)
					os.Exit(1)
				}
			} else {
				fmt.Println("error: must provide one of --url or --json")
				os.Exit(1)
			}

			fmt.Printf("Parsed %d events\n", len(rows))

			calendars := parse.Calendars(Year, rows)
			eventCount := 0

			for _, cal := range calendars {
				outpath := filepath.Join(OutDir, cal.Filename)
				out, err := os.Create(outpath)
				if err != nil {
					fmt.Printf("error writing calendar file %s: %v", outpath, err)
					os.Exit(1)
				}
				cal.WriteTo(out)
				out.Close()
				eventCount += cal.EmitCount
			}

			s, err := status.New(time.Now(), eventCount)
			if err != nil {
				fmt.Printf("failed to create status template: %v", err)
				os.Exit(1)
			}

			statusOut, err := os.Create(filepath.Join(OutDir, "status.html"))
			if err != nil {
				fmt.Printf("error opening file %s: %v", filepath.Join(OutDir, "status.html"), err)
				os.Exit(1)
			}
			defer statusOut.Close()
			s.WriteTo(statusOut)
		},
	}

	rootCmd.Flags().StringVarP(&OutDir, "outdir", "o", "", "directory in which to create or overwrite files. Defaults to CWD.")
	rootCmd.Flags().IntVarP(&Year, "year", "y", time.Now().Year(), "Year of the event. Defaults to current year. Example: 2022")
	rootCmd.Flags().StringVarP(&ScheduleURL, "url", "u", "", "URL to fetch schedule HTML from")
	rootCmd.Flags().StringVarP(&JSONFile, "json", "j", "", "JSON file to parse schedule from")

	rootCmd.Execute()
}
