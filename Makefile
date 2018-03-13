build:
	@go build -o openapi2jsonschema

buildsamples:
	@mkdir -p out
	@./openapi2jsonschema  -spec="sample/swagger2_flat-object.yaml"                 -go_constants  -block_additional_properties  -out="./out"
	@./openapi2jsonschema  -spec="sample/swagger2_referenced-object.yaml"           -go_constants  -block_additional_properties  -out="./out"
	@./openapi2jsonschema  -spec="sample/swagger2_object-with-pattern.yaml"         -go_constants  -block_additional_properties  -out="./out"
	@./openapi2jsonschema  -spec="sample/swagger2_array-of-referenced-object.yaml"  -go_constants  -block_additional_properties  -out="./out"

test:
	@go test
