# Proposal: Calendar View Using FullCalendar

## Goal

Add an interactive calendar view to the website so visitors can browse the SnF
schedule visually — by day and time — without needing to import anything into
external software first. This makes the site useful to a broader audience,
including people who just want a quick look at what's happening when.

## Library

[FullCalendar](https://fullcalendar.io/) (MIT license) renders interactive
calendar UIs in the browser. It has a free, fully open-source core that includes
the week and day timeline views most useful for a multi-day event like SnF. No
account, API key, or paid plan is required.

FullCalendar can load `.ics` files directly via its
[`@fullcalendar/icalendar`](https://fullcalendar.io/docs/icalendar) plugin, which
means the existing generated `.ics` files can be used as-is — no new backend
work needed.

## Approach

### New page: `website/calendar.html`

Add a new page alongside `index.html` and `details.html`. Add a "Calendar" link
to the `<nav>` in all three pages.

### Data source

Load each category's `.ics` file as a separate FullCalendar event source. Each
source gets a distinct color so forums, workshops, presentations, etc. are
visually distinct. The existing per-category `.ics` files serve double duty: they
remain importable, and now also feed the calendar view.

```
/ical/forums.ics          → color: #1a73e8 (blue)
/ical/workshops.ics       → color: #e67c00 (orange)
/ical/presentations.ics   → color: #0f9d58 (green)
/ical/facilities.ics      → color: #9e9e9e (grey)
/ical/other.ics           → color: #7b1fa2 (purple)
```

### Default view

Open on the first day of SnF for the current year. FullCalendar's `timeGrid`
week view fits well since the expo runs Monday–Sunday and events have specific
start/end times.

Provide day-column navigation so users can drill into a single day without
losing the week context.

The expo start date is determined at runtime by fetching `/ical/meta.json` (see
below) before the calendar is initialized. No hardcoded date in the HTML.

### Filtering

Render a row of checkboxes — one per category — above the calendar. Toggling a
checkbox shows or hides that event source. This is simpler than FullCalendar's
built-in filtering and requires no additional plugins.

### Styling

The page uses the same `simple.css` stylesheet as the rest of the site.
FullCalendar's own styles are loaded from its CDN bundle. No custom CSS
framework is needed beyond scoping FullCalendar's container to avoid conflicts
with `simple.css` resets.

## Implementation sketch

```html
<!-- website/calendar.html (excerpt) -->
<head>
  <link rel="stylesheet" href="https://cdn.simplecss.org/simple.min.css">
  <script src='https://cdn.jsdelivr.net/npm/fullcalendar@6/index.global.min.js'></script>
  <script src='https://cdn.jsdelivr.net/npm/@fullcalendar/icalendar@6/index.global.min.js'></script>
</head>
<body>
  ...
  <main>
    <div id="filters">
      <label><input type="checkbox" data-source="forums"        checked> Forums</label>
      <label><input type="checkbox" data-source="workshops"     checked> Workshops</label>
      <label><input type="checkbox" data-source="presentations" checked> Interviews &amp; Presentations</label>
      <label><input type="checkbox" data-source="facilities"    checked> Facilities</label>
      <label><input type="checkbox" data-source="other"         checked> Other</label>
    </div>
    <div id="calendar"></div>
  </main>

  <script>
    const sources = {
      forums:        { url: '/ical/forums.ics',        color: '#1a73e8' },
      workshops:     { url: '/ical/workshops.ics',     color: '#e67c00' },
      presentations: { url: '/ical/presentations.ics', color: '#0f9d58' },
      facilities:    { url: '/ical/facilities.ics',    color: '#9e9e9e' },
      other:         { url: '/ical/other.ics',         color: '#7b1fa2' },
    };

    const meta = await fetch('/ical/meta.json').then(r => r.json());

    const calendar = new FullCalendar.Calendar(document.getElementById('calendar'), {
      plugins: ['icalendar'],   // loaded via the global bundle
      initialView: 'timeGridWeek',
      initialDate: meta.expoStart,
      headerToolbar: {
        left:   'prev,next today',
        center: 'title',
        right:  'timeGridWeek,timeGridDay',
      },
      eventSources: Object.values(sources).map(s => ({
        url:    s.url,
        format: 'ics',
        color:  s.color,
      })),
    });
    calendar.render();

    // Checkbox filtering
    document.querySelectorAll('#filters input[type=checkbox]').forEach(cb => {
      cb.addEventListener('change', () => {
        const src = sources[cb.dataset.source];
        if (cb.checked) {
          calendar.addEventSource({ url: src.url, format: 'ics', color: src.color });
        } else {
          calendar.getEventSourceById(src.url)?.remove();
        }
      });
    });
  </script>
</body>
```

## meta.json

The Go tool already parses all event dates. After sorting rows into calendars,
it finds the minimum `DTSTART` across all events and writes `/ical/meta.json`
alongside the `.ics` and `status.html` files:

```json
{ "expoStart": "2026-04-13", "generatedAt": "2026-04-05T14:00:00Z" }
```

### Go changes

A new `meta` package (or an addition to the existing `status` package) handles
this. The pattern mirrors how `status.New()` works today:

```go
// In cmd/snf2ical.go, after calendars are written:
expoStart := earliestDate(calendars)
if err := meta.Write(filepath.Join(OutDir, "meta.json"), expoStart, time.Now()); err != nil {
    fmt.Printf("failed to write meta.json: %v\n", err)
    os.Exit(1)
}
```

`earliestDate` iterates the `[]Calendar` slice, scans each `Row`, parses the
`Start` field (already done by `timestamp()` in `parse.go`), and returns the
earliest date as a `YYYY-MM-DD` string.

### Benefits over a hardcoded date

- No per-year manual update required
- The date is authoritative — derived from the actual event data
- `generatedAt` gives the calendar page a place to show freshness information
  if desired in the future

## Nginx

The `.ics` files are already served from `/ical/`. No nginx changes are needed.
The existing `default_type text/calendar` location block in `nginx.conf` ensures
FullCalendar's iCalendar plugin receives the correct MIME type when it fetches
the files via XHR.

## What this does not require

- No new build steps — FullCalendar and its iCalendar plugin are loaded from CDN
- No server-side rendering or API endpoints
- No FullCalendar paid plan or license key
