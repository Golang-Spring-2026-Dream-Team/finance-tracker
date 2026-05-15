.PHONY: grpc

grpc:
	buf generate
	mv pb/proto/*.go pb/ 2>/dev/null || true
	rm -rf pb/proto pb/finance-tracker
