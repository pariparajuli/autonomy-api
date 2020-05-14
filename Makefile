.PHONY: api

dist =

default: build

api:
	go build -o bin/api main.go

score-worker:
	go build -o bin/score-worker background/command/score-worker/main.go

nudge-worker:
	go build -o bin/nudge-worker background/command/nudge-worker/main.go

run-api: api
	./bin/api -c config.yaml

run-score-worker: score-worker
	./bin/score-worker -c config.yaml

run-nudge-worker: nudge-worker
	./bin/nudge-worker -c config.yaml

bin: api score-worker nudge-worker

build-api-image:
ifndef dist
	$(error dist is undefined)
endif
	docker build --build-arg dist=$(dist) -t autonomy:api-$(dist) .
	docker tag autonomy:api-$(dist)  083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:api-$(dist)

build-score-worker-image:
ifndef dist
	$(error dist is undefined)
endif
	docker build --build-arg dist=$(dist) -t autonomy:score-worker-$(dist) . -f Dockerfile-ScoreWorker
	docker tag autonomy:score-worker-$(dist)  083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:score-worker-$(dist)

build-nudge-worker-image:
ifndef dist
	$(error dist is undefined)
endif
	docker build --build-arg dist=$(dist) -t autonomy:nudge-worker-$(dist) . -f Dockerfile-NudgeWorker
	docker tag autonomy:nudge-worker-$(dist)  083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:nudge-worker-$(dist)

build-crawler-image:
ifndef dist
	$(error dist is undefined)
endif
	docker build --build-arg dist=$(dist) -t autonomy:crawler-$(dist) . -f Dockerfile-Crawler
	docker tag autonomy:crawler-$(dist)  083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:crawler-$(dist)

push-worker:
ifndef dist
	$(error dist is undefined)
endif
	aws ecr get-login-password | docker login --username AWS --password-stdin 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com
	docker push 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:score-worker-$(dist)

push:
ifndef dist
	$(error dist is undefined)
endif
	aws ecr get-login-password | docker login --username AWS --password-stdin 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com
	docker push 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:api-$(dist)
	docker push 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:score-worker-$(dist)
	docker push 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:nudge-worker-$(dist)
	docker push 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/autonomy:crawler-$(dist)

build: build-api-image build-score-worker-image build-nudge-worker-image build-crawler-image

clean:
	rm -rf bin
