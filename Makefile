.PHONY: protos

protos:
	protoc service/service.proto --go_out=plugins=grpc:..
