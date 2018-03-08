package main

import (
	"testing"

	openapi2proto "github.com/NYTimes/openapi2proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GenerateJSONSchemas_WithEnums(t *testing.T) {
	api, err := openapi2proto.LoadDefinition("sample/swagger2_flat-object-with-enum.yaml")
	require.NoError(t, err)
	schemas, err := MapOpenAPIDefinitionsToJSONSchema(api)

	assert.NoError(t, err)
	assert.NotNil(t, schemas)
}
