package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Namespace string

const (
	NamespaceServer Namespace = "server"
	NamespaceACL    Namespace = "acl"
)

var stringToNamespace = map[string]Namespace{
	string(NamespaceServer): NamespaceServer,
	string(NamespaceACL):    NamespaceACL,
}

type Line struct {
	text string
	line int
}

type Config struct {
	lines map[Namespace][]Line
}

func ParseFile(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	c := Config{
		lines: make(map[Namespace][]Line, 0),
	}

	var line int
	for scanner.Scan() {
		line++

		fields := strings.Fields(scanner.Text())

		// ignore empty lines
		if len(fields) == 0 {
			continue
		}

		// ignore comments
		if len(fields) > 0 && len(fields[0]) > 0 && fields[0][0] == '#' {
			continue
		}

		if len(fields) < 2 {
			fmt.Fprintf(os.Stderr, "Config Error: parsing line %d: not enough arguments: '%s'\n", line, scanner.Text())
			continue
		}

		ns, ok := stringToNamespace[fields[0]]
		if !ok {
			fmt.Fprintf(os.Stderr, "Config Error: parsing line %d: '%s' is not a valid namespace.\n", line, fields[0])
			continue
		}

		if _, ok := c.lines[ns]; !ok {
			c.lines[ns] = make([]Line, 0)
		}

		c.lines[ns] = append(c.lines[ns], Line{
			text: strings.Join(fields[1:], " "),
			line: line,
		})

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &c, nil
}
