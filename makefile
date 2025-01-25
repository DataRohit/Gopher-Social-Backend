# 🛠️  Makefile for Docker Management 🐳

# 🌈 Colors and Emojis
GREEN  := $(shell tput setaf 2)
YELLOW := $(shell tput setaf 3)
BLUE   := $(shell tput setaf 4)
NC     := $(shell tput sgr0)
ROCKET := 🚀
TRASH  := 🗑️
CHECK  := ✅
INFO   := ℹ️
DOCS   := 📚
MIG    := ⚙️

# 📜 Help Command
.PHONY: help
help:
	@echo "${BLUE}Usage: make [command]${NC}"
	@echo ""
	@echo "${YELLOW}🐳  Docker Commands:${NC}"
	@echo "${GREEN}  docker-build${NC}        - ${ROCKET}	Build and start Docker containers in detached mode"
	@echo "${GREEN}  docker-clean${NC}        - ${TRASH}	Clean up Docker resources (images, containers, volumes, caches)"
	@echo ""
	@echo "${YELLOW}📚  Documentation Commands:${NC}"
	@echo "${GREEN}  gen-docs${NC}            - ${DOCS}	Generate Swagger documentation"
	@echo ""
	@echo "${YELLOW}⚙️  Migration Commands:${NC}"
	@echo "${GREEN}  migrate-create${NC}      - ${MIG}	Create a new database migration file"
	@echo "${GREEN}  migrate-up${NC}          - ${MIG}	Apply all pending migrations"
	@echo "${GREEN}  migrate-down${NC}        - ${MIG}	Roll back migrations (specify number of steps)"
	@echo "${GREEN}  migrate-clean${NC}       - ${MIG}	Clear all applied migrations from the database"
	@echo ""
	@echo "${YELLOW}🛠️   Utility Commands:${NC}"
	@echo "${GREEN}  help${NC}                - ${INFO}	Show this help message"

# 🚀 Build and Start Docker Containers
.PHONY: docker-build
docker-build:
	@echo "${GREEN}${ROCKET} Building and starting Docker containers...${NC}"
	docker compose up -d --build
	@echo "${GREEN}${CHECK} Docker containers are up and running!${NC}"

# 🗑️ Clean Docker Resources
.PHONY: docker-clean
docker-clean:
	@echo "${YELLOW}${TRASH} Cleaning up Docker resources...${NC}"
	@echo "${YELLOW}  - Stopping and removing containers...${NC}"
	docker compose down --remove-orphans
	@echo "${YELLOW}  - Removing unused images...${NC}"
	docker image prune -af
	@echo "${YELLOW}  - Removing unused volumes...${NC}"
	docker volume prune -f
	@echo "${YELLOW}  - Removing build cache...${NC}"
	docker builder prune -af
	@echo "${GREEN}${CHECK} Docker resources cleaned up!${NC}"

# 📚 Generate Swagger Documentation
.PHONY: gen-docs
gen-docs:
	@echo "${GREEN}${DOCS} Generating Swagger documentation...${NC}"
	swag init
	@echo "${GREEN}${CHECK} Swagger documentation generated!${NC}"

# ⚙️ Create a new database migration file
.PHONY: migrate-create
migrate-create:
	@echo "${BLUE}${MIG}  Enter migration file name (e.g., create_users_table):${NC}"
	@read -p "Migration file name: " file_name; \
	echo "${GREEN}${MIG}  Creating migration file: database/migrations/$${file_name}.sql ${NC}"; \
	migrate create -ext sql -dir database/migrations -seq $${file_name}
	@echo "${GREEN}${CHECK}  Migration file created!${NC}"

# ⚙️ Apply all pending migrations
.PHONY: migrate-up
migrate-up:
	@echo "${GREEN}${MIG} Applying all pending migrations...${NC}"
	migrate -database ${DATABASE_URL} -path database/migrations up
	@echo "${GREEN}${CHECK} Migrations applied successfully!${NC}"

# ⚙️ Roll back migrations (specify number of steps)
.PHONY: migrate-down
migrate-down:
	@echo "${BLUE}${MIG}  Enter the number of steps to roll back:${NC}"
	@read -p "Number of steps: " steps; \
	echo "${GREEN}${MIG}  Rolling back migrations...${NC}"; \
	migrate -database ${DATABASE_URL} -path database/migrations down $${steps}
	@echo "${GREEN}${CHECK} Migrations rolled back successfully!${NC}"

# ⚙️ Clear all applied migrations from the database
.PHONY: migrate-clean
migrate-clean:
	@echo "${YELLOW}${MIG}  Clearing all applied migrations from the database...${NC}"
	migrate -database ${DATABASE_URL} -path database/migrations drop
	@echo "${GREEN}${CHECK} All migrations cleared from the database!${NC}"