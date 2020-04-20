.PHONY: api background

dist =

default: build

api:
	go build -o bin/api main.go

score-worker:
	go build -o bin/score-worker background/command/score-worker/main.go

run-api: api
	./bin/api -c config.yaml

run-score-worker: score-worker
	./bin/score-worker -c config.yaml

bin: api score-worker

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

build-crawler-image:
ifndef dist
	$(error dist is undefined)
endif
	docker build --build-arg dist=$(dist) -t autonomy:crawler-$(dist) . -f Dockerfile-Crawler
	docker tag autonomy:crawler-$(dist)  083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:crawler-$(dist)

push-crawler:
ifndef dist
	$(error dist is undefined)
endif
	aws ecr get-login-password | docker login --username AWS --password-stdin 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com
	docker push 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:crawler-$(dist)

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
