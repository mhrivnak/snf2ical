package parse

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jordic/goics"
)

var months map[string]string = map[string]string{
	"March": "Mar",
	"April": "Apr",
}

type Event struct {
	Day      string `json:"day"` // "Wed. April 6"
	Name     string `json:"eventname"`
	Start    string `json:"starttime"` // "10:30:00 AM"
	End      string `json:"endtime"`
	Speaker  string `json:"speakernametitle"`
	Type     string `json:"eventtype"`
	Location string `json:"location"`
	Track    string `json:"track"`
}

type Row struct {
	Value Event `json:"value"`
}

type Calendar struct {
	Rows     []Row
	Filename string
	Name     string
}

func (c Calendar) Write(w io.Writer) {
	goics.NewICalEncode(w).Encode(c)
}

func Sorted(rows []Row) []Calendar {
	forum := Calendar{
		Name:     "SnF Forums",
		Filename: "forums.ics",
		Rows:     []Row{},
	}
	workshop := Calendar{
		Name:     "SnF Workshops",
		Filename: "workshops.ics",
		Rows:     []Row{},
	}
	other := Calendar{
		Name:     "SnF Other",
		Filename: "other.ics",
		Rows:     []Row{},
	}

	for _, row := range rows {
		switch row.Value.Type {
		case "Forum":
			forum.Rows = append(forum.Rows, row)
		case "Workshop", "SNF Workshop":
			workshop.Rows = append(workshop.Rows, row)
		default:
			other.Rows = append(other.Rows, row)
		}
	}

	return []Calendar{forum, workshop, other}
}

func (c Calendar) EmitICal() goics.Componenter {
	e := goics.NewComponent()
	e.SetType("VCALENDAR")
	e.AddProperty("PRODID", "github.com/mhrivnak/snf2ical")
	e.AddProperty("CALSCALE", "GREGORIAN")
	e.AddProperty("VERSION", "2.0")
	e.AddProperty("X-WR-TIMEZONE", "America/New_York")

	e.AddProperty("NAME", c.Name)
	e.AddProperty("X-WR-CALNAME", c.Name)

	for _, row := range c.Rows {
		// at least one entry was entirely blank, and another lacked a start or end time.
		if row.Value.Start != "" && row.Value.Name != "Forum - Not Currently Scheduled" {
			e.AddComponent(row.Value.AsICS())
		}
	}
	return e

}

func (e Event) AsICS() *goics.Component {
	c := goics.NewComponent()
	c.SetType("VEVENT")

	dayParts := strings.Split(e.Day, " ")
	if len(dayParts) != 3 {
		fmt.Printf("failed to parse Day field: %s\n", e.Day)
		os.Exit(1)
	}
	month, ok := months[dayParts[1]]
	if !ok {
		fmt.Printf("failed to parse Day field's month: %s\n", e.Day)
		os.Exit(1)
	}
	day, err := strconv.Atoi(dayParts[2])
	if err != nil {
		fmt.Printf("failed to parse Day field's day: %s\n", e.Day)
		os.Exit(1)
	}

	start, err := timestamp(month, day, e.Start)
	if err != nil {
		fmt.Printf("failed to parse start time \"%s\": %v\n", e.Start, err)
		os.Exit(1)
	}

	var end time.Time
	description := fmt.Sprintf("Type: %s\nTrack: %s\nSpeaker: %s", e.Type, e.Track, e.Speaker)
	// at least one event doesn't have an end time specified
	if e.End == "" {
		end = start.Add(time.Hour)
		description += "\n\nNOTE: no end time was specified."
	} else {
		end, err = timestamp(month, day, e.End)
		if err != nil {
			fmt.Printf("failed to parse end time \"%s\": %v\n", e.End, err)
			os.Exit(1)
		}
	}

	k, v := goics.FormatDateTimeField("DTSTART", start)
	c.AddProperty(k, v)
	k, v = goics.FormatDateTimeField("DTEND", end)
	c.AddProperty(k, v)
	c.AddProperty("SUMMARY", e.Name)
	c.AddProperty("LOCATION", e.Location)
	c.AddProperty("DESCRIPTION", description)

	return c
}

func timestamp(month string, day int, t string) (time.Time, error) {
	return time.Parse("Jan 2 2006 3:04:05 PM MST", fmt.Sprintf("%s %d 2022 %s EDT", month, day, t))
}
