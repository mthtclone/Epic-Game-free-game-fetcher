package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"time"
)

type Formatter struct {
	Title		string 		
	ProductSlug string 		
	ExpiryDate	time.Time	
	URL			string		
}

func FormatText(games []Formatter) string {
	var buf bytes.Buffer

	for _, game := range games {
		buf.WriteString(game.Title + "\n")
		buf.WriteString(game.URL + "\n")
		buf.WriteString("Free until: " + game.ExpiryDate.UTC().Format(time.RFC1123) + "\n\n")
	}

	return buf.String()
}

func FormatJSON(games []Formatter) ([]byte, error) {
	return json.MarshalIndent(games, "", "  ")
}

func FormatHTML(games []Formatter) (string, error) {
	const tpl = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Free Epic Games</title>
	<style>
		body { font-family: sans-serif; padding: 3rem;}
		.game { margin-bottom: 20px; }
	</style>
</head>
<body>
	{{ range . }}
	<div class="game">
		<h2>{{ .Title }}</h2>
		<a href="{{ .URL }}">{{ .URL }}</a><br>
		<small>Free until: {{ .ExpiryDate.Format "Mon, 02 Jan 2006 15:04 MST" }}</small>
	</div>
	{{ else }}
	<p>No free games right now.</p>
	{{ end }}
</body>
</html>
`

	t, err := template.New("html").Parse(tpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, games); err != nil {
		return "", err
	}

	return buf.String(), nil
}

