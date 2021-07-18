package main

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type Project struct {
	Root   string   `json:"root"`
	Bucket string   `json:"bucket"`
	Prefix string   `json:"prefix"`
	Ignore []string `json:"ignore"`
}

func LoadProject(path string) (*Project, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p := &Project{}
	err = yaml.NewDecoder(f).Decode(p)
	if err != nil {
		return nil, err
	}

	err = p.validate()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Project) validate() error {
	if p.Root == "" {
		return fmt.Errorf("root is not specified")
	}

	if p.Bucket == "" {
		return fmt.Errorf("bucket is not specified")
	}

	return nil
}
