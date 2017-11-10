build:
	@go build -o openapi2jsonschema

test:
	@mkdir -p out
	@./openapi2jsonschema -block_additional_properties -spec="sample/swagger2-flat.yaml" -out="./out"
	@./openapi2jsonschema -spec="sample/swagger2-references.yaml" -out="./out"
	@./openapi2jsonschema -spec="sample/swagger2-arrays-of-references.yaml" -out="./out"
