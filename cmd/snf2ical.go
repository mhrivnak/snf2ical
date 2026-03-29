package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mhrivnak/snf2ical/pkg/parse"
	"github.com/mhrivnak/snf2ical/pkg/status"

	"github.com/spf13/cobra"
)

var OutDir string
var Year int
var ScheduleURL string

func main() {
	rootCmd := &cobra.Command{
		Use:   "snf2ical [input-file]",
		Short: "generate ical files for Sun n Fun",
		Long:  "snf2ical creates or overwrites several ical files for Sun n Fun",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var rows []parse.Row
			var err error

			// If URL is provided, fetch from URL; otherwise read from file
			if ScheduleURL != "" {
				fmt.Printf("Fetching schedule from %s...\n", ScheduleURL)
				rows, err = parse.FetchScheduleHTML(ScheduleURL)
				if err != nil {
					fmt.Printf("failed to fetch schedule from URL: %v\n", err)
					os.Exit(1)
				}
			} else {
				if len(args) == 0 {
					fmt.Println("error: either provide an input file or use --url flag")
					os.Exit(1)
				}

				inputpath := args[0]
				f, err := os.Open(inputpath)
				if err != nil {
					fmt.Println("failed to open file")
					os.Exit(1)
				}
				defer f.Close()

				data, err := ioutil.ReadAll(f)
				if err != nil {
					fmt.Println("failed to parse data")
					os.Exit(1)
				}

				// Try to parse as HTML first, fall back to JSON
				if strings.Contains(inputpath, ".html") || strings.Contains(string(data), "<table") {
					f.Seek(0, 0) // Reset file pointer
					rows, err = parse.ParseScheduleHTML(f)
					if err != nil {
						fmt.Printf("failed to parse HTML: %v\n", err)
						os.Exit(1)
					}
				} else {
					rows = make([]parse.Row, 0)
					if err = json.Unmarshal(data, &rows); err != nil {
						fmt.Printf("failed to unmarshal json file: %v", err)
						os.Exit(1)
					}
				}
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
	rootCmd.Flags().IntVarP(&Year, "year", "y", 0, "Year of the event. Example: 2022")
	rootCmd.Flags().StringVarP(&ScheduleURL, "url", "u", "", "URL to fetch schedule HTML from")
	rootCmd.MarkFlagRequired("year")

	rootCmd.Execute()
}
