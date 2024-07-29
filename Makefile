
all: protos up


ifeq ($(OS), Windows_NT)
make-dir:
	@if not exist $(DIR) mkdir $(DIR)
else
make-dir:
	mkdir -p $(DIR)
endif


# PROTOS SECTION
.PHONY: go-protos py-protos protos


go-protos:
	$(MAKE) make-dir DIR=Backend\Go-api\proto_generated
	protoc \
		--go_out=./Backend/Go-api/proto_generated \
		--go_opt=paths=source_relative \
		--go-grpc_out=./Backend/Go-api/proto_generated \
		--go-grpc_opt=paths=source_relative \
		--proto_path=proto \
		proto/ram_generator.proto
	cd Backend/Go-api && go mod tidy

ifeq ($(OS),Windows_NT)
fix-py-generated-imports:
	powershell -Command "(Get-Content Backend\Python-ai\ai_generator\proto_generated\ram_generator_pb2_grpc.py) -replace '^import (\S+_pb2) as (\S+__pb2)', 'from . import $$1 as $$2' | Set-Content Backend\Python-ai\ai_generator\proto_generated\ram_generator_pb2_grpc.py"
else
fix-py-generated-imports:
	cd Backend/Python-ai/ai_generator/proto_generated && sed -i 's/^import .*_pb2 as/from . \0/' ram_generator_pb2_grpc.py
endif

py-protos:
	$(MAKE) make-dir DIR=Backend\Python-ai\ai_generator\proto_generated

	python -m grpc_tools.protoc proto/ram_generator.proto \
		--python_out=Backend/Python-ai/ai_generator/proto_generated \
		--grpc_python_out=Backend/Python-ai/ai_generator/proto_generated \
		--pyi_out=Backend/Python-ai/ai_generator/proto_generated \
		--proto_path=proto
# Костыль  для фикса относительных импортов в сгенерированных файлах
	$(MAKE) fix-py-generated-imports


protos: go-protos py-protos

# DOCKER SECTION
.PHONY: build, up, start, stop, down, restart
SERVICE?=

%:
	@:

build:
	docker-compose build $(filter-out $@,$(MAKECMDGOALS))
up:
	docker-compose up -d $(filter-out $@,$(MAKECMDGOALS))
start:
	docker-compose start $(filter-out $@,$(MAKECMDGOALS))
stop:
	docker-compose stop $(filter-out $@,$(MAKECMDGOALS))
down:
	docker-compose down $(filter-out $@,$(MAKECMDGOALS))
restart:
	docker-compose restart $(filter-out $@,$(MAKECMDGOALS))