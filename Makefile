.PHONY: api background

dist =

default: build

api:
	go build -o bin/api main.go

background:
	go build -o bin/background background/background/main.go

run-api: api
	./bin/api -c config.yaml

run-background: background
	./bin/background -c config.yaml

all: api background

build-api-image:
ifndef dist
	$(error dist is undefined)
endif
	docker build --build-arg dist=$(dist) -t autonomy:api-$(dist) .
	docker tag autonomy:api-$(dist)  083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:api-$(dist)

build-background-image:
ifndef dist
	$(error dist is undefined)
endif
	docker build --build-arg dist=$(dist) -t autonomy:background-$(dist) . -f Dockerfile-Job
	docker tag autonomy:background-$(dist)  083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:background-$(dist)

push:
ifndef dist
	$(error dist is undefined)
endif
	aws ecr get-login-password | docker login --username AWS --password-stdin 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com
	docker push 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:api-$(dist)
	docker push 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:background-$(dist)

build: build-api-image build-background-image

clean:
	rm -rf bin
