MODULE     := github.com/apranto/pgoperator
BINARY     := pgoperator
KIND_CLUSTER := pgoperator-dev

# Go settings
GOFLAGS    := -v
LDFLAGS    := -w -s

.PHONY: all build run test clean cluster cluster-delete deploy undeploy crd sample fmt vet

## ---- Build ----

all: fmt vet build

build:
	go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o bin/$(BINARY) .

run:
	go run . --kubeconfig=$(HOME)/.kube/config

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./... -v

clean:
	rm -rf bin/

## ---- Kind Cluster ----

cluster:
	kind create cluster --config hack/kind-config.yaml
	@echo "Cluster '$(KIND_CLUSTER)' created. Context set automatically."
	kubectl cluster-info --context kind-$(KIND_CLUSTER)

cluster-delete:
	kind delete cluster --name $(KIND_CLUSTER)

## ---- Deploy to cluster ----

crd:
	kubectl apply -f deploy/crds/

sample:
	kubectl apply -f deploy/samples/sample-postgresdb.yaml

deploy: crd
	kubectl apply -f deploy/rbac/
	kubectl apply -f deploy/operator/

undeploy:
	kubectl delete -f deploy/operator/ --ignore-not-found
	kubectl delete -f deploy/rbac/ --ignore-not-found
	kubectl delete -f deploy/crds/ --ignore-not-found

## ---- Helpers ----

# Watch PostgresDB resources
watch:
	kubectl get pgdb -w

# Describe a specific PostgresDB
describe:
	kubectl describe pgdb my-postgres
