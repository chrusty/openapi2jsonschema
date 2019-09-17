package filewriter

import (
	"testing"

	"github.com/chrusty/openapi2jsonschema/pkg/schemaconverter/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDeriveGoConstantsFilename(t *testing.T) {
	schemaWriter := New(&types.Config{
		GoConstants:         true,
		GoConstantsFilename: "constants",
		OutPath:             "/output/schemas",
	}, logrus.New())

	filename := schemaWriter.deriveGoConstantsFilename("cruft")
	assert.Equal(t, "/output/schemas/constantsCruft.go", filename)
}

func TestDeriveJSONSchemaFilename(t *testing.T) {
	schemaWriter := New(&types.Config{
		JSONSchemaFileExtention: "jsonschema",
		OutPath:                 "/output/schemas",
	}, logrus.New())

	filename := schemaWriter.deriveJSONSchemaFilename("cruft/cruft")
	assert.Equal(t, "/output/schemas/cruft/cruft.jsonschema", filename)
}

func TestDeriveSpecPathFilename(t *testing.T) {
	schemaWriter := New(&types.Config{
		JSONSchemaFileExtention: "jsonschema",
		SpecPath:                "/input/spec/openapi.yaml",
	}, logrus.New())

	filename := schemaWriter.deriveSpecPathFilename()
	assert.Equal(t, "openapi", filename)
}

func TestWriteToFile(t *testing.T) {
	schemaWriter := New(&types.Config{}, logrus.New())

	assert.NoError(t, schemaWriter.writeToFile("/tmp/cruft", []byte("cruft")))
	assert.Error(t, schemaWriter.writeToFile("/cruft/cruft.cft", []byte("cruft")))
}
