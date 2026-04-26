.PHONY: up down reset demo reseed logs ps help

help:
	@echo "ClearFly developer shortcuts"
	@echo ""
	@echo "  make up       — start the full stack in the background"
	@echo "  make down     — stop the stack but keep volumes"
	@echo "  make reset    — stop the stack and wipe the database (fresh demo data)"
	@echo "  make demo     — populate the running stack with realistic demo data"
	@echo "  make reseed   — reset + start + demo (one-shot clean demo)"
	@echo "  make logs     — follow logs from all services"
	@echo "  make ps       — show container status"

up:
	cd docker && docker compose up -d --build

down:
	cd docker && docker compose down

reset:
	cd docker && docker compose down -v

demo:
	python3 scripts/seed_demo.py

reseed: reset up
	@echo "Waiting for gateway to become healthy…"
	@sleep 12
	$(MAKE) demo

logs:
	cd docker && docker compose logs -f

ps:
	cd docker && docker compose ps
