package types

// Converter turns Swagger / OpenAPI specs into JSONSchemas:
type Converter interface {
	GenerateJSONSchemas() ([]GeneratedJSONSchema, error)
}
