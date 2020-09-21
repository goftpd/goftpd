package config

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/goftpd/goftpd/ftp"
	"github.com/pkg/errors"
)

func (c *Config) ParseServerOpts() (*ftp.ServerOpts, error) {
	var opts ftp.ServerOpts

	if _, ok := c.lines[NamespaceServer]; !ok {
		return nil, errors.New("no server options provided")
	}

	rv := reflect.ValueOf(&opts)

	for _, l := range c.lines[NamespaceServer] {
		fields := strings.Fields(l.text)

		if len(fields) < 2 {
			return nil, errors.Errorf("error on line %d", l.line)
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
								return nil, errors.Errorf("error parsing int on line %d: too many fields", l.line)
							}

							num, err := strconv.Atoi(fields[1])
							if err != nil {
								return nil, errors.Errorf("error parsing int on line %d: not a number", l.line)
							}

							reflect.Indirect(rv).Field(i).SetInt(int64(num))

						case reflect.Slice:
							switch reflect.Indirect(rv).Field(i).Type().Elem().Kind() {
							case reflect.Int:

								var nums []int

								for _, f := range fields[1:] {
									num, err := strconv.Atoi(f)
									if err != nil {
										return nil, errors.Errorf("error parsing int on line %d: not a number", l.line)
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

	return &opts, nil
}
