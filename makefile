go-protos:
	cd Backend && protoc --go_out=./Go/proto_generated --go_opt=paths=source_relative --go-grpc_out=./Go/proto_generated --go-grpc_opt=paths=source_relative --proto_path=proto proto/ram_generator.proto
py-protos:
	cd Backend && python -m grpc_tools.protoc --python_out=Python/ai_generator --grpc_python_out=Python/ai_generator --pyi_out=Python/ai_generator --proto_path=proto  proto/ram_generator.proto
protos:
	cd Backend && protoc --go_out=./Go/proto_generated --go_opt=paths=source_relative --go-grpc_out=./Go/proto_generated --go-grpc_opt=paths=source_relative --proto_path=proto proto/ram_generator.proto && python -m grpc_tools.protoc --python_out=Python/ai_generator --grpc_python_out=Python/ai_generator --pyi_out=Python/ai_generator --proto_path=proto  proto/ram_generator.proto
