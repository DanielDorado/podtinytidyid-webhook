# Go build
.PHONY: build
build:
	mkdir -p ./bin
	go build -o ./bin/podtinytidyid

.PHONY: test
test:
	go test

.PHONY: docker-build
docker-build:
	scripts/docker-build.sh

.PHONY: manifest-build
manifest-build:
	scripts/k8s-build.sh

.PHONY: manifest-clean
manifest-clean:
	rm build -rf

.PHONY: k8s-deploy
k8s-deploy:
	kubectl get namespace/podtinytidyid-webhook > /dev/null || ( sleep 4 && kubectl create namespace podtinytidyid-webhook )
	kubectl apply -f build

.PHONY:
 k8s-clean:
	( kubectl get MutatingWebhookConfiguration/podtinytidyid-webhook && kubectl delete ValidatingWebhookConfiguration/podtinytidyid-webhook ) || true
	kubectl delete namespace/podtinytidyid-webhook || true
