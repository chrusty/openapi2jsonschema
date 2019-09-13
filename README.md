# openapi2jsonschema
Produce JSONSchemas from OpenAPI2 (Swagger) & OpenAPI3 definitions

## Features
* Supports **OpenAPI2** (Swagger) and **OpenAPI3** (with the `-v3` flag)
* Creates a JSONSchema for each model within the provided spec, and writes each to its own file
* Optionally generates an importable GoLang package containing constants for each JSONSchema (in case you want to have access to the JSONSchemas from code without having to deal with loading files)

## Usage:
```
Usage of bin/openapi2jsonschema:
  -allow_null_values
    	Allow NULL values as well as the defined types?
  -block_additional_properties
    	Block additional properties?
  -go_constants
    	Output GoLang constants (in addition to JSONSchemas)?
  -loglevel string
    	Log level [trace, debug, info, warn, error] (default "info")
  -out string
    	Where to write jsonschema output files to (default "./out")
  -spec string
    	Location of the swagger spec file (default "spec.yaml")
  -v3
    	Use OpenAPI3 (instead of Swagger 2)?
```
