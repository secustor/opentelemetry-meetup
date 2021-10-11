
KIND_ADDITIONAL_ARGS?=--kubeconfig /home/secustor/.kube/config
KIND_CLUSTER_NAME?=kind

# prepare setup
create-kind-cluster:
	kind create cluster ${KIND_ADDITIONAL_ARGS} --name ${KIND_CLUSTER_NAME}  --config=./hack/kind_config.yaml

set-context:
	kubectl config set-context kind-${KIND_CLUSTER_NAME}

deploy-ingress:
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

deploy-kafka:
	helmsman -f ./deploy/setup.yaml -apply
	kubectl apply -f ./deploy/kafka_cluster.yaml

# teardown setup
delete-kind-cluster:
	kind delete cluster --name ${KIND_CLUSTER_NAME}

