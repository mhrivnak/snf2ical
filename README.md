# snf2ical

snf2ical is a tool that converts the raw event data from Sun 'n Fun into
iCalendar files that can be imported into calendar software.

The Sun 'n Fun online schedule is basically unusable due to how it's rendered in
a table that's incapable of sorting chronologically. Thus I found it useful to
write this tool that ingests the event data and creates ical files.

## Output files

The tool writes one `.ics` file per event category plus supporting files:

| File | Contents |
|------|----------|
| `forums.ics` | Forum sessions |
| `workshops.ics` | Workshops |
| `presentations.ics` | Presentations and interviews |
| `facilities.ics` | Parking, tickets, merchandise, family zone, seating |
| `other.ics` | Everything else |
| `meta.json` | Expo start date and generation timestamp |
| `status.html` | Human-readable generation status page |

## Usage

```
snf2ical [flags]

Flags:
  -u, --url string      URL to fetch the schedule HTML from
  -j, --json string     JSON file to parse the schedule from (alternative to --url)
  -o, --outdir string   Directory to write output files into (default: current directory)
  -y, --year int        Year of the event (default: current year)
  -h, --help            Show help
```

Either `--url` or `--json` must be provided.

### Fetch live schedule and write files to a directory

```sh
snf2ical --url https://dailyschedule.flysnf.org/dailyschedule-table/ --outdir /var/www/snf2ical/ical/
```

### Parse from a previously saved JSON file

```sh
snf2ical --json sched.json --outdir ./output/
```

## Building

Requires Go 1.25 or later.

```sh
go build -o snf2ical ./cmd/snf2ical.go
```

Or use the Taskfile:

```sh
task build
```

## Running periodically

The iCalendar files should be regenerated regularly so that calendar
subscribers receive schedule updates. A simple cron job works well:

```cron
0 * * * * /usr/local/bin/snf2ical --url https://dailyschedule.flysnf.org/dailyschedule-table/ --outdir /var/www/snf2ical/ical/
```

## Serving the website

The `website/` directory contains the static site. An example nginx
configuration is provided in `nginx.conf`. Copy it to
`/etc/nginx/conf.d/`, update the domain name and certificate paths,
copy the `website/` contents and generated `.ics` files to the configured
root, and reload nginx.

## Project structure

```
cmd/            CLI entry point
pkg/parse/      HTML and iCalendar parsing
pkg/meta/       meta.json generation
pkg/status/     status.html generation
website/        Static site (index, calendar view, details page)
enhancements/   Design proposals for future features
nginx.conf      Example nginx TLS configuration
Taskfile.yml    Build, test, and dev tasks
```
