package parse

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

// FetchScheduleHTML fetches and parses the schedule from the given URL
func FetchScheduleHTML(url string) ([]Row, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return ParseScheduleHTML(resp.Body)
}

// ParseScheduleHTML parses the HTML table from the reader
func ParseScheduleHTML(r io.Reader) ([]Row, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var rows []Row
	tableFound := false
	var f func(*html.Node)
	f = func(n *html.Node) {
		// Short-circuit if table already found
		if tableFound {
			return
		}
		if n.Type == html.ElementNode && n.Data == "table" {
			// Check if this is the schedule table
			if hasClass(n, "tablepress") {
				rows = parseTable(n)
				tableFound = true
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if !tableFound {
		return nil, fmt.Errorf("no table found in HTML")
	}

	return rows, nil
}

// hasClass checks if a node has a specific class (exact token match)
func hasClass(n *html.Node, class string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			// Split class attribute on whitespace and check for exact match
			classes := strings.Fields(attr.Val)
			for _, c := range classes {
				if c == class {
					return true
				}
			}
		}
	}
	return false
}

// parseTable extracts rows from a table node
func parseTable(table *html.Node) []Row {
	rows := []Row{}
	var tbody *html.Node

	// Find tbody element
	for c := table.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tbody" {
			tbody = c
			break
		}
	}

	if tbody == nil {
		return rows
	}

	// Parse each row in tbody
	for tr := tbody.FirstChild; tr != nil; tr = tr.NextSibling {
		if tr.Type == html.ElementNode && tr.Data == "tr" {
			row := parseRow(tr)
			// Skip empty rows (rows where all fields are empty)
			if !isEmptyEvent(row.Value) {
				rows = append(rows, row)
			}
		}
	}

	return rows
}

// parseRow extracts data from a table row
func parseRow(tr *html.Node) Row {
	cells := []string{}

	for td := tr.FirstChild; td != nil; td = td.NextSibling {
		if td.Type == html.ElementNode && td.Data == "td" {
			cells = append(cells, getTextContent(td))
		}
	}

	// Map cells to Event struct fields in order:
	// Day, Event Name, Start Time, End Time, Speaker Name/Title, Event Type, Tracks & Credits, Location
	event := Event{}
	if len(cells) > 0 {
		event.Day = strings.TrimSpace(cells[0])
	}
	if len(cells) > 1 {
		event.Name = strings.TrimSpace(cells[1])
	}
	if len(cells) > 2 {
		event.Start = strings.TrimSpace(cells[2])
	}
	if len(cells) > 3 {
		event.End = strings.TrimSpace(cells[3])
	}
	if len(cells) > 4 {
		event.Speaker = strings.TrimSpace(cells[4])
	}
	if len(cells) > 5 {
		event.Type = strings.TrimSpace(cells[5])
	}
	if len(cells) > 6 {
		event.Track = strings.TrimSpace(cells[6])
	}
	if len(cells) > 7 {
		event.Location = strings.TrimSpace(cells[7])
	}

	return Row{Value: event}
}

// getTextContent extracts all text from a node and its children
func getTextContent(n *html.Node) string {
	var text strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return text.String()
}

// isEmptyEvent checks if an event has all empty fields (after trimming whitespace)
func isEmptyEvent(e Event) bool {
	return strings.TrimSpace(e.Day) == "" && strings.TrimSpace(e.Name) == "" &&
		strings.TrimSpace(e.Start) == "" && strings.TrimSpace(e.End) == "" &&
		strings.TrimSpace(e.Speaker) == "" && strings.TrimSpace(e.Type) == "" &&
		strings.TrimSpace(e.Track) == "" && strings.TrimSpace(e.Location) == ""
}
