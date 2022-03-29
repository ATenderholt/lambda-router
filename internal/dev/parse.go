package dev

import (
	"github.com/ATenderholt/lambda-router/internal/domain"
	"gopkg.in/yaml.v2"
	"io"
	"os"
)

func Parse(reader io.Reader) (map[string]domain.DevFunction, error) {
	results := make(map[string]domain.DevFunction)
	decoder := yaml.NewDecoder(reader)
	err := decoder.Decode(&results)
	if err != nil {
		return nil, err
	}

	// Decorate results with name
	for key, value := range results {
		v := value
		v.SetName("dev-" + key)
		results[key] = v
	}

	return results, nil
}

func ParseFile(filename string) (map[string]domain.DevFunction, error) {
	f, err := os.Open(filename)
	if err != nil {
		logger.Errorf("Unable to open %s: %v", filename, err)
		return nil, err
	}
	defer f.Close()

	return Parse(f)
}
