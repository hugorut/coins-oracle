package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
)

var (
	files = []string{
		"gettransaction.json",
		"getinfo.json",
		"getbalance.json",
	}
)

type templateParams struct {
	Snake   string
	Name    string
	AssetID string

	TemplateDir string
	ClientDir   string
	FixtureDir  string

	help bool
}

// NameUpper returns name in upper case.
func (t templateParams) NameUpper() string {
	return strings.ToUpper(t.Name)
}

func main() {
	tp := parseFlags()
	if tp.help {
		flag.Usage()
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	tp.Snake = strings.ToLower(tp.Snake)
	tp.Name = strings.Title(camelCase(tp.Snake, false))
	tp.AssetID = strings.ToUpper(tp.AssetID)

	templateDir := path.Join(dir, tp.TemplateDir)
	clientDir := path.Join(dir, tp.ClientDir)

	client, err := template.ParseFiles(path.Join(templateDir, "client.tmpl"))
	if err != nil {
		log.Fatal(err)
	}

	clientTest, err := template.ParseFiles(path.Join(templateDir, "client_test.tmpl"))
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(path.Join(clientDir, "client_"+tp.Snake+".go"))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Execute(f, tp)
	if err != nil {
		log.Fatal(err)
	}

	f, err = os.Create(path.Join(clientDir, "client_"+tp.Snake+"_test.go"))
	err = clientTest.Execute(f, tp)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(path.Join(tp.FixtureDir, tp.Snake, "req"), 0775)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(path.Join(tp.FixtureDir, tp.Snake, "res"), 0775)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		createFixture(tp, "req", file)
		createFixture(tp, "res", file)
	}

	appendClient(tp)
}

func parseFlags() templateParams {
	var tp templateParams

	flag.BoolVar(&tp.help, "help", false, "print the cmd options and defaults")

	flag.StringVar(&tp.Snake, "name", "coin", "Snake cased name of used for the coin being generated. e.g. bitcoin_sv")
	flag.StringVar(&tp.AssetID, "asset_id", "c", "The asset id for the coin being generated.")
	flag.StringVar(&tp.TemplateDir, "tmpl_dir", "./internal/transport/tmpl", "the directory of the templates to parse")
	flag.StringVar(&tp.ClientDir, "client_dir", "./internal/transport", "the directory of the output files")
	flag.StringVar(&tp.FixtureDir, "fixture_dir", "./internal/test/fixtures", "the directory to output the fixture files")

	flag.Parse()

	return tp
}

func appendClient(tp templateParams) {
	resolver := path.Join(tp.ClientDir, "resolver.go")
	input, err := ioutil.ReadFile(resolver)
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(input), "\n")

	var lastRegister int
	newLines := make([]string, len(lines)+1)

	for i, line := range lines {
		if strings.Contains(line, "r.Register") {
			lastRegister = i
		}
		newLines[i] = line
	}

	copy(newLines[lastRegister+1:], newLines[lastRegister:])
	newLines[lastRegister+1] = "	r.Register(" + tp.Name + "AssetID, must(New" + tp.Name + "Client()))"

	output := strings.Join(newLines, "\n")
	err = ioutil.WriteFile(resolver, []byte(output), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func createFixture(tp templateParams, dir, file string) {
	f, err := os.Create(path.Join(tp.FixtureDir, tp.Snake, dir, file))
	if err != nil {
		log.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// toLower converts a character in the range of ASCII characters 'A' to 'Z' to its lower
// case counterpart. Other characters remain the same.
func toLower(ch rune) rune {
	if ch >= 'A' && ch <= 'Z' {
		return ch + 32
	}
	return ch
}

// toLower converts a character in the range of ASCII characters 'a' to 'z' to its lower
// case counterpart. Other characters remain the same.
func toUpper(ch rune) rune {
	if ch >= 'a' && ch <= 'z' {
		return ch - 32
	}
	return ch
}

// isSpace checks if a character is some kind of whitespace.
func isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// isDelimiter checks if a character is some kind of whitespace or '_' or '-'.
func isDelimiter(ch rune) bool {
	return ch == '-' || ch == '_' || isSpace(ch)
}

func camelCase(s string, upper bool) string {
	s = strings.TrimSpace(s)
	buffer := make([]rune, 0, len(s))

	var prev rune
	for _, curr := range s {
		if !isDelimiter(curr) {
			if isDelimiter(prev) || (upper && prev == 0) {
				buffer = append(buffer, toUpper(curr))
			} else {
				buffer = append(buffer, toLower(curr))
			}
		}
		prev = curr
	}

	return string(buffer)
}
