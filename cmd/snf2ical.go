package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/mhrivnak/snf2ical/pkg/parse"
	"github.com/mhrivnak/snf2ical/pkg/status"

	"github.com/spf13/cobra"
)

var OutDir string
var Year int

func main() {
	rootCmd := &cobra.Command{
		Use:   "snf2ical",
		Short: "generate ical files for Sun n Fun",
		Long:  "snf2ical creates or overwrites several ical files for Sun n Fun",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
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

			rows := make([]parse.Row, 0)

			if err = json.Unmarshal(data, &rows); err != nil {
				fmt.Printf("failed to unmarshal json file: %v", err)
				os.Exit(1)
			}

			s, err := status.New(time.Now(), len(rows))
			if err != nil {
				fmt.Printf("failed to create status template: %v", err)
				os.Exit(1)
			}

			calendars := parse.Calendars(Year, rows)

			for _, cal := range calendars {
				outpath := filepath.Join(OutDir, cal.Filename)
				out, err := os.Create(outpath)
				if err != nil {
					fmt.Printf("error writing calendar file %s: %v", outpath, err)
					os.Exit(1)
				}
				cal.Write(out)
				out.Close()
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
	rootCmd.MarkFlagRequired("year")

	rootCmd.Execute()
}
