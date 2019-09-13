package oapi3

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/types"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Converter performs schema conversion:
type Converter struct {
	config                     *types.Config
	logger                     *logrus.Logger
	nestedAdditionalProperties map[string]json.RawMessage
	swagger                    *openapi3.Swagger
}

// New takes a config and returns a new Converter:
func New(config *types.Config, logger *logrus.Logger) (*Converter, error) {

	// Load the OpenAPI spec:
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile(config.SpecPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to load spec (%s)", config.SpecPath)
	}

	// Make sure the provided spec is really OpenAPI 3.x:
	if !strings.HasPrefix(swagger.OpenAPI, "3") {
		return nil, fmt.Errorf("This spec (%s) is not OpenAPI 3.x", swagger.OpenAPI)
	}

	logger.WithField("title", swagger.Info.Title).WithField("version", swagger.Info.Version).Info("Ready to convert Swagger / OpenAPI3")
	logger.WithField("description", swagger.Info.Description).Trace("Description")

	// Return a new *Converter:
	return &Converter{
		config:  config,
		logger:  logger,
		swagger: swagger,
	}, nil
}

// GenerateJSONSchemas takes an OpenAPI "Spec" and converts each definition into a JSONSchema:
func (c *Converter) GenerateJSONSchemas() ([]types.GeneratedJSONSchema, error) {

	c.logger.Debug("Converting API")

	// Store the output in here:
	generatedJSONSchemas, err := c.mapOpenAPIDefinitionsToJSONSchema()
	if err != nil {
		return nil, errors.Wrap(err, "could not map openapi definitions to jsonschema")
	}

	return generatedJSONSchemas, nil
}
