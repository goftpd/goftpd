package config

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Namespace string

const (
	NamespaceVar    Namespace = "var"
	NamespaceServer Namespace = "server"
	NamespaceACL    Namespace = "acl"
	NamespaceFS     Namespace = "fs"
	NamespaceAuth   Namespace = "auth"
)

var stringToNamespace = map[string]Namespace{
	string(NamespaceServer): NamespaceServer,
	string(NamespaceACL):    NamespaceACL,
	string(NamespaceFS):     NamespaceFS,
	string(NamespaceVar):    NamespaceVar,
	string(NamespaceAuth):   NamespaceAuth,
}

type Line struct {
	text string
	line int
}

type Config struct {
	lines map[Namespace][]Line

	variables map[string]string
}

func ParseFile(file string) (*Config, error) {
	c := Config{
		lines:     make(map[Namespace][]Line, 0),
		variables: make(map[string]string, 0),
	}

	// first read in any variables
	if err := c.parseVariables(file); err != nil {
		return nil, err
	}

	// then read everything else
	if err := c.parseLines(file); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Config) parseVariables(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

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

		// check if this is a variable
		if ns != NamespaceVar {
			continue
		}
		c.variables[fields[1]] = strings.Join(fields[2:], " ")
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (c *Config) parseLines(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

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

		// check if this is a variable
		if ns == NamespaceVar {
			continue
		}

		for idx, f := range fields {
			if len(f) > 1 && f[0] == '$' {
				v, ok := c.variables[f[1:]]
				if !ok {
					return fmt.Errorf("Config Error: uninitialized variable '%s' on line %d", f, line)
				}
				fields[idx] = v
			}
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
		return err
	}

	return nil
}

// parse searches opts for any fields tagged with "goftpd" and then attempts
// to parse and insert the corresponding Line
func (c *Config) parse(lines []Line, opts interface{}) error {
	rv := reflect.ValueOf(opts)

	for _, l := range lines {
		fields := strings.Fields(l.text)

		if len(fields) < 2 {
			return errors.Errorf("error on line %d", l.line)
		}

		tag := strings.ToLower(fields[0])

		// iterate over the opts
		for i := 0; i < reflect.Indirect(rv).NumField(); i++ {

			if value, ok := reflect.Indirect(rv).Type().Field(i).Tag.Lookup("goftpd"); ok {
				if value == tag {
					if reflect.Indirect(rv).Field(i).CanSet() {
						switch reflect.Indirect(rv).Field(i).Kind() {

						case reflect.String:
							reflect.Indirect(rv).Field(i).SetString(strings.Join(fields[1:], " "))

						case reflect.Int:
							if len(fields) > 2 {
								return errors.Errorf("error parsing int on line %d: too many fields", l.line)
							}

							num, err := strconv.Atoi(fields[1])
							if err != nil {
								return errors.Errorf("error parsing int on line %d: not a number", l.line)
							}

							reflect.Indirect(rv).Field(i).SetInt(int64(num))

						case reflect.Slice:
							switch reflect.Indirect(rv).Field(i).Type().Elem().Kind() {
							case reflect.Int:

								var nums []int

								for _, f := range fields[1:] {
									num, err := strconv.Atoi(f)
									if err != nil {
										return errors.Errorf("error parsing int on line %d: not a number", l.line)
									}
									nums = append(nums, num)
								}

								reflect.Indirect(rv).Field(i).Set(reflect.ValueOf(nums))
							}
						}
					}
				}
			}
		}
	}

	return nil
}
