package dev_test

import (
	"github.com/ATenderholt/rainbow-functions/internal/dev"
	"github.com/ATenderholt/rainbow-functions/internal/domain"
	"strings"
	"testing"
)

const example = `---
func1:
  handler: main.handler
  runtime: python3.8
  basePath: a/relative/path
func2:
  handler: main.other_handler
  runtime: python3.6
  basePath: /an/absolute/path
  environment:
    - KEY=VALUE
    - FOO=BAR
`

func TestExample(t *testing.T) {
	reader := strings.NewReader(example)
	results, err := dev.Parse(reader)
	if err != nil {
		t.Errorf("Unable to parse yaml: %v", err)
		t.FailNow()
	}

	dev1, ok := results["func1"]
	if !ok {
		t.Errorf("func1 not found")
	}

	assertThat(t, dev1).
		HasName("dev-func1").
		HasHandler("main.handler").
		HasRuntime("python3.8").
		HasBasePath("a/relative/path").
		HasEnvironment([]string{})

	dev2, ok := results["func2"]
	if !ok {
		t.Errorf("func2 not found")
	}

	assertThat(t, dev2).
		HasName("dev-func2").
		HasHandler("main.other_handler").
		HasRuntime("python3.6").
		HasBasePath("/an/absolute/path").
		HasEnvironment([]string{"KEY=VALUE", "FOO=BAR"})
}

type DevFunctionAssertion struct {
	t      *testing.T
	actual domain.DevFunction
}

func assertThat(t *testing.T, actual domain.DevFunction) *DevFunctionAssertion {
	return &DevFunctionAssertion{t: t, actual: actual}
}

func (d *DevFunctionAssertion) HasName(name string) *DevFunctionAssertion {
	if d.actual.Name() != name {
		d.t.Errorf("Expected name %s, but has %s", name, d.actual.Name())
		d.t.Fail()
	}

	return d
}

func (d *DevFunctionAssertion) HasHandler(handler string) *DevFunctionAssertion {
	if d.actual.Handler != handler {
		d.t.Errorf("Expected handler %s, but has %s", handler, d.actual.Handler)
		d.t.Fail()
	}

	return d
}

func (d *DevFunctionAssertion) HasRuntime(runtime string) *DevFunctionAssertion {
	if d.actual.Runtime != runtime {
		d.t.Errorf("Expected runtime %s, but has %s", runtime, d.actual.Runtime)
		d.t.Fail()
	}

	return d
}

func (d *DevFunctionAssertion) HasBasePath(basePath string) *DevFunctionAssertion {
	if d.actual.BasePath != basePath {
		d.t.Errorf("Expected basePath %s, but has %s", basePath, d.actual.BasePath)
		d.t.Fail()
	}

	return d
}

func (d *DevFunctionAssertion) HasEnvironment(environment []string) *DevFunctionAssertion {
	actualSize := len(d.actual.Environment)
	size := len(environment)
	if actualSize != size {
		d.t.Errorf("Expected environment to have %d values, but has only %d", size, actualSize)
	}

	for i, value := range environment {
		actualValue := d.actual.Environment[i]
		if actualValue != value {
			d.t.Errorf("Expected environment value %d to be %s, but is %s", i, value, actualValue)
		}
	}

	return d
}
