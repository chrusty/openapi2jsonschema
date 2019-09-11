package schemaconverter

import (
	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/filewriter"
	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/openapi2"
	"github.com/chrusty/openapi2jsonschema/internal/schemaconverter/types"

	"github.com/sirupsen/logrus"
)

// NewV2 returns an OpenAPIv2 schema converter:
func NewV2(config *types.Config, logger *logrus.Logger) (*openapi2.Converter, error) {
	return openapi2.New(config, logger)
}

// NewWriter returns a schema writer:
func NewWriter(config *types.Config, logger *logrus.Logger) *filewriter.Writer {
	return filewriter.New(config, logger)
}
