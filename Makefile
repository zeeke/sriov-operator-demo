TARGET_BIN=$(PWD)/bin/sriov-operator-demo
KUBECONFIG?=$(PWD)/bin/kubeconfig

build: $(TARGET_BIN)
$(TARGET_BIN):
	go build -o $(TARGET_BIN)

generate-examples:
	./scripts/generate-examples.sh

setup-kind: $(KUBECONFIG)
$(KUBECONFIG):
	./scripts/setup-kind.sh

update-doc:
	./scripts/update-doc.sh

.PHONY: build generate-examples setup-kind update-doc

