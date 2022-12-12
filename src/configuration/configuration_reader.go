package configuration

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type fieldDescriptor struct {
	realField   reflect.Value
	typeOfField reflect.StructField
	envVarName  string
}

func getFieldDescriptorsOf(spec any) []fieldDescriptor {
	target := reflect.ValueOf(spec).Elem()
	typeOfTarget := target.Type()

	var fields []fieldDescriptor

	for i := 0; i < typeOfTarget.NumField(); i++ {
		field := typeOfTarget.Field(i)

		varName := field.Tag.Get("env")
		if varName == "" {
			varName = strings.ToUpper(field.Name)
		}

		descriptor := fieldDescriptor{
			realField:   target.Field(i),
			typeOfField: field,
			envVarName:  varName,
		}

		fields = append(fields, descriptor)
	}

	return fields
}

func (p *reader) LoadConfigurationInto(cfg any) {
	fields := getFieldDescriptorsOf(cfg)
	for _, field := range fields {
		for _, source := range p.sources {
			envVarValue := source.Get(field.envVarName)
			if envVarValue != "" {
				switch field.realField.Type().Kind() {
				case reflect.String:
					field.realField.SetString(envVarValue)
				case reflect.Bool:
					value := false
					if envVarValue == "1" || strings.EqualFold("true", envVarValue) {
						value = true
					}
					field.realField.SetBool(value)
				default:
					panic(fmt.Sprintf(
						"Type \"%s\" has field \"%s\" with unsupported type of \"%s\"",
						reflect.ValueOf(cfg).Elem().Type().Name(),
						field.typeOfField.Name,
						field.typeOfField.Type.Kind(),
					))
				}
			}
		}
	}
}

type reader struct {
	sources []ValueSource
}

type ConfigurationReader interface {
	LoadConfigurationInto(cfg any)
}
type ValueSource interface {
	Get(key string) string
}

// region value sources

type osValueSource struct {
}

func (o *osValueSource) Get(key string) string {
	return os.Getenv(key)
}

type inMemoryValueSource struct {
	values map[string]string
}

func (o *inMemoryValueSource) Get(key string) string {
	return o.values[key]
}

// endregion

func readLinesFromFile(filePath string) []string {
	var result []string

	file, err := os.Open(filePath)
	if err != nil {
		return result
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		result = append(result, line)
	}

	return result
}

func newValueSourceFrom(lines []string) ValueSource {
	envVars := make(map[string]string)
	for _, line := range lines {
		before, after, wasFound := strings.Cut(line, "=")
		if wasFound {
			envVars[before] = after
		}
	}

	return &inMemoryValueSource{values: envVars}
}

func NewConfigurationReader() ConfigurationReader {
	return &reader{
		sources: []ValueSource{
			newValueSourceFrom(readLinesFromFile("./.env")),
			&osValueSource{},
		},
	}
}

func LoadInto(cfg any) {
	NewConfigurationReader().LoadConfigurationInto(cfg)
}

//func NewConfigurationReaderWithSources(sources ...ValueSource) ConfigurationReader {
//	return &reader{sources: sources}
//}
