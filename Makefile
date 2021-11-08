OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

BUILD_CMD?=docker

KIND_ADDITIONAL_ARGS?=--kubeconfig ~/.kube/config
KIND_CLUSTER_NAME?=kind

# Images
APP_SNAPSHOT_IMAGE?=localhost/snapshot:test
APP_PRODUCER_IMAGE?=localhost/producer:test
APP_CONSUMER_IMAGE?=localhost/consumer:test


# prepare setup
create-kind-cluster:
	kind create cluster ${KIND_ADDITIONAL_ARGS} --name ${KIND_CLUSTER_NAME}  --config=./hack/kind_config.yaml

set-context:
	kubectl config set-context kind-${KIND_CLUSTER_NAME}

prepare-environment: deploy-setup deploy-kafka

deploy-setup:
	helmsman -f ./deploy/setup.yaml -apply

deploy-kafka:
	kubectl apply -f ./deploy/kafka_cluster.yaml

# teardown setup
delete-kind-cluster:
	kind delete cluster --name ${KIND_CLUSTER_NAME}


prepare-apps: build-all load-all

# Build applications
build-all: build-snapshot build-producer build-consumer

build-snapshot:
	$(BUILD_CMD) build -t ${APP_SNAPSHOT_IMAGE} ./apps/snapshot

build-producer:
	$(BUILD_CMD) build --platform=$(OS)/$(ARCH) --build-arg CMD_BIN=cmd/producer.go -t ${APP_PRODUCER_IMAGE} apps/report

build-consumer:
	$(BUILD_CMD) build --platform=$(OS)/$(ARCH) --build-arg CMD_BIN=cmd/consumer.go -t ${APP_CONSUMER_IMAGE} apps/report

# Load images
load-all: load-snapshot load-producer load-consumer

load-snapshot:
	kind load docker-image ${APP_SNAPSHOT_IMAGE}

load-producer:
	kind load docker-image ${APP_PRODUCER_IMAGE}

load-consumer:
	kind load docker-image ${APP_CONSUMER_IMAGE}
