package status

import (
	"io"
	"text/template"
	"time"
)

type status struct {
	Updated   string
	NumEvents int
	Template  *template.Template
}

type Status interface {
	WriteTo(w io.Writer) error
}

func New(updated time.Time, numEvents int) (Status, error) {
	tmpl, err := template.New("status").Parse(t)
	if err != nil {
		return nil, err
	}

	return status{
		Updated:   updated.String(),
		NumEvents: numEvents,
		Template:  tmpl,
	}, nil
}

func (s status) WriteTo(w io.Writer) error {
	return s.Template.Execute(w, s)
}

const t = `<!DOCTYPE html>
<html lang="en">
<head>
    <link rel="stylesheet" href="https://cdn.simplecss.org/simple.min.css">
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Status</title>
</head>
<body>
    <main>
	    <p>Last re-generated: {{ .Updated }}</p>
		<p>Number of events: {{ .NumEvents }}</p>
	</main>
</body>
</html>
`
