up: ## start
	docker-compose up

down: ## stop
	docker-compose down

restart: | down up ## start & stop

in: ## in
	docker-compose exec gcs sh

run: ## run
	go run .

runw: ## run waste
	go run . --waste

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
