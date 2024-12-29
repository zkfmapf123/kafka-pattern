up:
	docker-compose up --build -d

cdc-up:
	docker-compose up -f docker-compose.cdc.yml up --build -d

outbox-up:
	docker-compose up -f docker-compose.outbox.yml up --build -d

down:	
	docker-compose down
