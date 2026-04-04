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
	var thead *html.Node
	var tbody *html.Node

	// Find thead and tbody elements
	for c := table.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			if c.Data == "thead" {
				thead = c
			} else if c.Data == "tbody" {
				tbody = c
			}
		}
	}

	if tbody == nil {
		return rows
	}

	// Extract column mapping from headers
	columnMap := extractColumnMap(thead)

	// Parse each row in tbody
	for tr := tbody.FirstChild; tr != nil; tr = tr.NextSibling {
		if tr.Type == html.ElementNode && tr.Data == "tr" {
			row := parseRow(tr, columnMap)
			// Skip empty rows (rows where all fields are empty)
			if !isEmptyEvent(row.Value) {
				rows = append(rows, row)
			}
		}
	}

	return rows
}

// extractColumnMap creates a mapping from expected column names to column indices
func extractColumnMap(thead *html.Node) map[string]int {
	columnMap := make(map[string]int)

	if thead == nil {
		return columnMap
	}

	// Find the header row
	var headerRow *html.Node
	for tr := thead.FirstChild; tr != nil; tr = tr.NextSibling {
		if tr.Type == html.ElementNode && tr.Data == "tr" {
			headerRow = tr
			break
		}
	}

	if headerRow == nil {
		return columnMap
	}

	// Extract header text and build mapping
	colIndex := 0
	for th := headerRow.FirstChild; th != nil; th = th.NextSibling {
		if th.Type == html.ElementNode && th.Data == "th" {
			headerText := strings.TrimSpace(getTextContent(th))
			columnMap[headerText] = colIndex
			colIndex++
		}
	}

	return columnMap
}

// parseRow extracts data from a table row using column mapping
func parseRow(tr *html.Node, columnMap map[string]int) Row {
	cells := []string{}

	for td := tr.FirstChild; td != nil; td = td.NextSibling {
		if td.Type == html.ElementNode && td.Data == "td" {
			cells = append(cells, getTextContent(td))
		}
	}

	// Map cells to Event struct fields using header names
	event := Event{}

	// Helper function to get cell value by header name
	getCellByHeader := func(headerName string) string {
		if idx, ok := columnMap[headerName]; ok && idx < len(cells) {
			return strings.TrimSpace(cells[idx])
		}
		return ""
	}

	event.Day = getCellByHeader("Day")
	event.Name = getCellByHeader("Event Name")
	event.Start = getCellByHeader("Start Time")
	event.End = getCellByHeader("End Time")
	event.Speaker = getCellByHeader("Speaker Name/Title")
	event.Type = getCellByHeader("Event Type")
	event.Track = getCellByHeader("Tracks & Credits")
	event.Location = getCellByHeader("Location")

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
