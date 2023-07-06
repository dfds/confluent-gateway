package configuration

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestConfiguration struct {
	Foo string `env:"TEST_FOO"`
	Bar string
	Baz bool `env:"TEST_BAZ"`
}

func TestReturnsExpectedWhenReadingTagsFromStruct(t *testing.T) {
	var cfg = TestConfiguration{}
	descriptor := getFieldDescriptorsOf(&cfg)

	assert.Equal(t, "TEST_FOO", descriptor[0].envVarName)
	assert.Equal(t, "BAR", descriptor[1].envVarName)
}

func TestReader_LoadConfigurationInto_InsertsExpectedValuesFromValueSource(t *testing.T) {
	reader := reader{sources: []ValueSource{
		&inMemoryValueSource{map[string]string{"TEST_FOO": "REAL_FOO_VALUE", "TEST_BAZ": "1"}},
	}}

	var cfg = TestConfiguration{}
	reader.LoadConfigurationInto(&cfg)

	assert.Equal(t, "REAL_FOO_VALUE", cfg.Foo)
	assert.Equal(t, "", cfg.Bar)
	assert.Equal(t, true, cfg.Baz)
}

func TestReader_newValueSourceFromHandlesValuesContainingEqualSignsWithGrace(t *testing.T) {
	sut := newValueSourceFrom([]string{"FOO=1=1, 2=2"})
	assert.Equal(t, "1=1, 2=2", sut.Get("FOO"))
}
