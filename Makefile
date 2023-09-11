docker-central:
	docker-compose -f docker-compose.yml up valve

docker-regional:
	docker-compose -f docker-compose.yml up regionales1 regionales2 regionales3 regionales4

docker-rabbit:
	docker-compose -f docker-compose.yml up rabbitmq

docker-down:
	docker-compose -f docker-compose.yml down

docker-clean:
	docker system prune -a