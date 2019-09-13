package types

// Writer handles writing JSONSchemas and Go constants to files:
type Writer interface {
	WriteJSONSchemasToFiles(generatedJSONSchemas []GeneratedJSONSchema) error
	WriteGoConstantsToFile(generatedJSONSchemas []GeneratedJSONSchema) error
}
