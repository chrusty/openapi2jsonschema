build:
	@go build -o openapi2jsonschema

test:
	@mkdir -p out
	# @./openapi2jsonschema -debug -block_additional_properties -spec="sample/swagger2-flat.yaml" -out="./out"
	@./openapi2jsonschema -debug -spec="sample/swagger2-nested.yaml" -out="./out"
	# @./openapi2jsonschema -debug -spec="sample/swagger2-array-of-refs.yaml" -out="./out"
