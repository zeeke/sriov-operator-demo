TARGET_BIN=$(PWD)/bin/sriov-operator-demo
KUBECONFIG?=$(PWD)/bin/kind_kubeconfig

build $(TARGET_BIN): 
	go build -o $(TARGET_BIN)

generate-examples:
	KUBECONFIG=$(KUBECONFIG) ./scripts/generate-examples.sh

setup-kind:
	./scripts/setup-kind.sh

update-doc:
	./scripts/update-doc.sh

.PHONY: build generate-examples setup-kind update-doc

