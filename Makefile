
KIND_ADDITIONAL_ARGS?=--kubeconfig ~/.kube/config
KIND_CLUSTER_NAME?=kind

# Images
APP_SNAPSHOT_IMAGE?=localhost/snapshot:test


# prepare setup
create-kind-cluster:
	kind create cluster ${KIND_ADDITIONAL_ARGS} --name ${KIND_CLUSTER_NAME}  --config=./hack/kind_config.yaml

set-context:
	kubectl config set-context kind-${KIND_CLUSTER_NAME}

deploy-setup:
	helmsman -f ./deploy/setup.yaml -apply

deploy-kafka:
	kubectl apply -f ./deploy/kafka_cluster.yaml

# teardown setup
delete-kind-cluster:
	kind delete cluster --name ${KIND_CLUSTER_NAME}


prepare-apps: build-all load-all

# Build applications
build-all: build-snapshot

build-snapshot:
	buildah build -t ${APP_SNAPSHOT_IMAGE} ./apps/snapshot



# Load images
load-all: load-snapshot

load-snapshot:
	kind load docker-image ${APP_SNAPSHOT_IMAGE}
