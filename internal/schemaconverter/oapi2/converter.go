package oapi2

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/types"

	openapi2proto "github.com/NYTimes/openapi2proto/openapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Converter performs schema conversion:
type Converter struct {
	config                     *types.Config
	logger                     *logrus.Logger
	nestedAdditionalProperties map[string]json.RawMessage
	spec                       *openapi2proto.Spec
}

// New takes a config and returns a new Converter:
func New(config *types.Config, logger *logrus.Logger) (*Converter, error) {

	// Load the OpenAPI spec:
	spec, err := openapi2proto.LoadFile(config.SpecPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to load spec (%s)", config.SpecPath)
	}

	// Make sure the provided spec is really OpenAPI 2.x:
	if !strings.HasPrefix(spec.Swagger, "2") {
		return nil, fmt.Errorf("This spec (%s) is not OpenAPI 2.x", spec.Swagger)
	}

	logger.WithField("title", spec.Info.Title).WithField("version", spec.Info.Version).Info("Ready to convert Swagger / OpenAPI2")
	logger.WithField("description", spec.Info.Description).Trace("Description")

	// Return a new *Converter:
	return &Converter{
		spec:                       spec,
		config:                     config,
		logger:                     logger,
		nestedAdditionalProperties: make(map[string]json.RawMessage),
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
