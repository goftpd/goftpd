package config

import (
	"github.com/goftpd/goftpd/script"
)

func (c *Config) ParseScripts() (script.Engine, error) {
	lines, ok := c.lines[NamespaceScript]
	if !ok {
		return &script.DummyEngine{}, nil
	}

	l := make([]string, len(lines))
	for i := range lines {
		l[i] = lines[i].text
	}

	engine, err := script.NewLUAEngine(l)
	if err != nil {
		return nil, err
	}

	return engine, nil
}
