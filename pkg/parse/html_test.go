package parse

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestParseScheduleHTML(t *testing.T) {
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
<table id="tablepress-5" class="tablepress tablepress-id-5">
<thead>
<tr class="row-1">
	<th class="column-1">Day</th><th class="column-2">Event Name</th><th class="column-3">Start Time</th><th class="column-4">End Time</th><th class="column-5">Speaker Name/Title</th><th class="column-6">Event Type</th><th class="column-7">Tracks &amp; Credits</th><th class="column-8">Location</th>
</tr>
</thead>
<tbody class="row-striping row-hover">
<tr class="row-2">
	<td class="column-1">Tue. April 7</td><td class="column-2">Camper Registration</td><td class="column-3">12:00 PM</td><td class="column-4">6:00 PM</td><td class="column-5"></td><td class="column-6">Guest Services</td><td class="column-7"></td><td class="column-8">Hamilton Rd. Entrance</td>
</tr>
<tr class="row-3">
	<td class="column-1">Wed. April 8</td><td class="column-2">Workshop: Introduction to Aviation</td><td class="column-3">10:00 AM</td><td class="column-4">12:00 PM</td><td class="column-5">John Doe, CFI</td><td class="column-6">Workshop</td><td class="column-7">Aviation 101</td><td class="column-8">Building A</td>
</tr>
<tr class="row-4">
	<td class="column-1"></td><td class="column-2"></td><td class="column-3"></td><td class="column-4"></td><td class="column-5"></td><td class="column-6"></td><td class="column-7"></td><td class="column-8"></td>
</tr>
<tr class="row-5">
	<td class="column-1">Thu. April 9</td><td class="column-2">Field Opens for Aircraft Arrivals</td><td class="column-3">12:00 PM</td><td class="column-4"></td><td class="column-5">NOTAM not in effect until Monday</td><td class="column-6">Arrivals</td><td class="column-7"></td><td class="column-8">Event Site</td>
</tr>
</tbody>
</table>
</body>
</html>
`

	rows, err := ParseScheduleHTML(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("ParseScheduleHTML failed: %v", err)
	}

	// Should have 3 rows (empty row should be skipped)
	if len(rows) != 3 {
		t.Fatalf("Expected 3 rows, got %d", len(rows))
	}

	// Test first row
	if rows[0].Value.Day != "Tue. April 7" {
		t.Errorf("Expected Day 'Tue. April 7', got '%s'", rows[0].Value.Day)
	}
	if rows[0].Value.Name != "Camper Registration" {
		t.Errorf("Expected Name 'Camper Registration', got '%s'", rows[0].Value.Name)
	}
	if rows[0].Value.Start != "12:00 PM" {
		t.Errorf("Expected Start '12:00 PM', got '%s'", rows[0].Value.Start)
	}
	if rows[0].Value.End != "6:00 PM" {
		t.Errorf("Expected End '6:00 PM', got '%s'", rows[0].Value.End)
	}
	if rows[0].Value.Type != "Guest Services" {
		t.Errorf("Expected Type 'Guest Services', got '%s'", rows[0].Value.Type)
	}
	if rows[0].Value.Location != "Hamilton Rd. Entrance" {
		t.Errorf("Expected Location 'Hamilton Rd. Entrance', got '%s'", rows[0].Value.Location)
	}

	// Test second row with speaker and track
	if rows[1].Value.Day != "Wed. April 8" {
		t.Errorf("Expected Day 'Wed. April 8', got '%s'", rows[1].Value.Day)
	}
	if rows[1].Value.Speaker != "John Doe, CFI" {
		t.Errorf("Expected Speaker 'John Doe, CFI', got '%s'", rows[1].Value.Speaker)
	}
	if rows[1].Value.Track != "Aviation 101" {
		t.Errorf("Expected Track 'Aviation 101', got '%s'", rows[1].Value.Track)
	}

	// Test third row with missing end time
	if rows[2].Value.Day != "Thu. April 9" {
		t.Errorf("Expected Day 'Thu. April 9', got '%s'", rows[2].Value.Day)
	}
	if rows[2].Value.End != "" {
		t.Errorf("Expected empty End time, got '%s'", rows[2].Value.End)
	}
}

func TestParseScheduleHTML_EmptyRows(t *testing.T) {
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
<table id="tablepress-5" class="tablepress tablepress-id-5">
<thead>
<tr class="row-1">
	<th class="column-1">Day</th><th class="column-2">Event Name</th><th class="column-3">Start Time</th><th class="column-4">End Time</th><th class="column-5">Speaker Name/Title</th><th class="column-6">Event Type</th><th class="column-7">Tracks &amp; Credits</th><th class="column-8">Location</th>
</tr>
</thead>
<tbody>
<tr>
	<td></td><td></td><td></td><td></td><td></td><td></td><td></td><td></td>
</tr>
<tr>
	<td>   </td><td>  </td><td>  </td><td></td><td></td><td></td><td></td><td></td>
</tr>
</tbody>
</table>
</body>
</html>
`

	rows, err := ParseScheduleHTML(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("ParseScheduleHTML failed: %v", err)
	}

	// All rows are empty and should be skipped
	if len(rows) != 0 {
		t.Errorf("Expected 0 rows, got %d", len(rows))
	}
}

func TestParseScheduleHTML_NoTable(t *testing.T) {
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
<p>No table here</p>
</body>
</html>
`

	_, err := ParseScheduleHTML(strings.NewReader(htmlContent))
	if err == nil {
		t.Error("Expected error when no table found, got nil")
	}
}

func TestParseScheduleHTML_ExactClassMatch(t *testing.T) {
	// Table with class "tablepressfoo" should NOT match "tablepress"
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
<table class="tablepressfoo">
<tbody>
<tr>
	<td>Should not match</td>
</tr>
</tbody>
</table>
</body>
</html>
`

	_, err := ParseScheduleHTML(strings.NewReader(htmlContent))
	if err == nil {
		t.Error("Expected error when table class is substring match only, got nil")
	}

	// Table with class "foo tablepress bar" SHOULD match "tablepress"
	htmlContent2 := `
<!DOCTYPE html>
<html>
<body>
<table class="foo tablepress bar">
<thead>
<tr><th>Day</th><th>Event Name</th><th>Start Time</th><th>End Time</th><th>Speaker Name/Title</th><th>Event Type</th><th>Tracks</th><th>Location</th></tr>
</thead>
<tbody>
<tr>
	<td>Mon. April 1</td><td>Test</td><td>10:00 AM</td><td>11:00 AM</td><td></td><td>Workshop</td><td></td><td>Building A</td>
</tr>
</tbody>
</table>
</body>
</html>
`

	rows, err := ParseScheduleHTML(strings.NewReader(htmlContent2))
	if err != nil {
		t.Errorf("Expected success with exact class token match, got error: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(rows))
	}
}

func TestParseScheduleHTML_StopsAfterFirstTable(t *testing.T) {
	// Multiple tables with tablepress class - should only parse the first one
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
<table class="tablepress">
<thead>
<tr><th>Day</th><th>Event Name</th><th>Start Time</th><th>End Time</th><th>Speaker Name/Title</th><th>Event Type</th><th>Tracks</th><th>Location</th></tr>
</thead>
<tbody>
<tr>
	<td>Mon. April 1</td><td>First Table Event</td><td>10:00 AM</td><td>11:00 AM</td><td></td><td>Workshop</td><td></td><td>Building A</td>
</tr>
</tbody>
</table>
<table class="tablepress">
<thead>
<tr><th>Day</th><th>Event Name</th><th>Start Time</th><th>End Time</th><th>Speaker Name/Title</th><th>Event Type</th><th>Tracks</th><th>Location</th></tr>
</thead>
<tbody>
<tr>
	<td>Tue. April 2</td><td>Second Table Event</td><td>2:00 PM</td><td>3:00 PM</td><td></td><td>Workshop</td><td></td><td>Building B</td>
</tr>
</tbody>
</table>
</body>
</html>
`

	rows, err := ParseScheduleHTML(strings.NewReader(htmlContent))
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("Expected 1 row from first table only, got %d", len(rows))
	}
	if len(rows) > 0 && rows[0].Value.Name != "First Table Event" {
		t.Errorf("Expected first table event, got %s", rows[0].Value.Name)
	}
}

func TestIsEmptyEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected bool
	}{
		{
			name:     "completely empty event",
			event:    Event{},
			expected: true,
		},
		{
			name:     "event with only day",
			event:    Event{Day: "Mon. April 1"},
			expected: false,
		},
		{
			name:     "event with only whitespace",
			event:    Event{Day: "   ", Name: "  "},
			expected: true,
		},
		{
			name: "fully populated event",
			event: Event{
				Day:      "Mon. April 1",
				Name:     "Test Event",
				Start:    "10:00 AM",
				End:      "11:00 AM",
				Speaker:  "Speaker Name",
				Type:     "Workshop",
				Track:    "Track 1",
				Location: "Building A",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmptyEvent(tt.event)
			if result != tt.expected {
				t.Errorf("isEmptyEvent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetTextContent(t *testing.T) {
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
<td>Simple text</td>
</body>
</html>
`

	doc, _ := parseHTMLFragment(htmlContent)
	text := getTextContent(doc)

	// The text should include "Simple text" somewhere
	if !strings.Contains(text, "Simple text") {
		t.Errorf("Expected text to contain 'Simple text', got '%s'", text)
	}
}

// Helper function to parse HTML fragment for testing
func parseHTMLFragment(content string) (*html.Node, error) {
	return html.Parse(strings.NewReader(content))
}

func TestParseScheduleHTML_ReorderedColumns(t *testing.T) {
	// Test with columns in different order than expected
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
<table class="tablepress">
<thead>
<tr><th>Location</th><th>Event Type</th><th>Day</th><th>Event Name</th><th>End Time</th><th>Start Time</th><th>Speaker Name/Title</th><th>Tracks & Credits</th></tr>
</thead>
<tbody>
<tr>
	<td>Building A</td><td>Workshop</td><td>Wed. April 8</td><td>Test Event</td><td>12:00 PM</td><td>10:00 AM</td><td>John Doe</td><td>Track 1</td>
</tr>
</tbody>
</table>
</body>
</html>
`

	rows, err := ParseScheduleHTML(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("ParseScheduleHTML failed: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	row := rows[0].Value
	if row.Day != "Wed. April 8" {
		t.Errorf("Expected Day 'Wed. April 8', got '%s'", row.Day)
	}
	if row.Name != "Test Event" {
		t.Errorf("Expected Name 'Test Event', got '%s'", row.Name)
	}
	if row.Start != "10:00 AM" {
		t.Errorf("Expected Start '10:00 AM', got '%s'", row.Start)
	}
	if row.End != "12:00 PM" {
		t.Errorf("Expected End '12:00 PM', got '%s'", row.End)
	}
	if row.Speaker != "John Doe" {
		t.Errorf("Expected Speaker 'John Doe', got '%s'", row.Speaker)
	}
	if row.Type != "Workshop" {
		t.Errorf("Expected Type 'Workshop', got '%s'", row.Type)
	}
	if row.Track != "Track 1" {
		t.Errorf("Expected Track 'Track 1', got '%s'", row.Track)
	}
	if row.Location != "Building A" {
		t.Errorf("Expected Location 'Building A', got '%s'", row.Location)
	}
}

func TestParseScheduleHTML_ExtraColumns(t *testing.T) {
	// Test with extra columns that aren't mapped
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
<table class="tablepress">
<thead>
<tr><th>Day</th><th>Extra Column 1</th><th>Event Name</th><th>Start Time</th><th>Extra Column 2</th><th>End Time</th><th>Speaker Name/Title</th><th>Event Type</th><th>Tracks & Credits</th><th>Location</th><th>Extra Column 3</th></tr>
</thead>
<tbody>
<tr>
	<td>Mon. April 1</td><td>Ignored</td><td>Test Event</td><td>9:00 AM</td><td>Ignored</td><td>10:00 AM</td><td>Jane Smith</td><td>Forum</td><td>Aviation</td><td>Hall B</td><td>Ignored</td>
</tr>
</tbody>
</table>
</body>
</html>
`

	rows, err := ParseScheduleHTML(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("ParseScheduleHTML failed: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	row := rows[0].Value
	if row.Day != "Mon. April 1" {
		t.Errorf("Expected Day 'Mon. April 1', got '%s'", row.Day)
	}
	if row.Name != "Test Event" {
		t.Errorf("Expected Name 'Test Event', got '%s'", row.Name)
	}
	if row.Location != "Hall B" {
		t.Errorf("Expected Location 'Hall B', got '%s'", row.Location)
	}
}
