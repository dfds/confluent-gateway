package configuration

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type Configuration struct {
	Foo string `env:"TEST_FOO"`
	Bar string
	Baz bool `env:"TEST_BAZ"`
}

func TestReturnsExpectedWhenReadingTagsFromStruct(t *testing.T) {
	var cfg = Configuration{}
	descriptor := getFieldDescriptorsOf(&cfg)

	assert.Equal(t, "TEST_FOO", descriptor[0].envVarName)
	assert.Equal(t, "BAR", descriptor[1].envVarName)
}

func TestReader_LoadConfigurationInto_InsertsExpectedValuesFromValueSource(t *testing.T) {
	reader := reader{sources: []ValueSource{
		&inMemoryValueSource{map[string]string{"TEST_FOO": "REAL_FOO_VALUE", "TEST_BAZ": "1"}},
	}}

	var cfg = Configuration{}
	reader.LoadConfigurationInto(&cfg)

	assert.Equal(t, "REAL_FOO_VALUE", cfg.Foo)
	assert.Equal(t, "", cfg.Bar)
	assert.Equal(t, true, cfg.Baz)
}