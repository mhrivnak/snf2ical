package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mhrivnak/snf2ical/pkg/parse"

	"github.com/jordic/goics"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "snf2ical",
		Short: "generate ical for Sun n Fun",
		Long:  "snf2ical generates an ical file for Sun n Fun",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filename := args[0]
			f, err := os.Open(filename)
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

			rows := make(parse.Rows, 1)

			if err = json.Unmarshal(data, &rows); err != nil {
				fmt.Printf("failed to unmarshal json file: %v", err)
				os.Exit(1)
			}

			goics.NewICalEncode(os.Stdout).Encode(rows)
		},
	}

	rootCmd.Execute()
}
