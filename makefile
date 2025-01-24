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

# 📜 Help Command
.PHONY: help
help:
	@echo "${BLUE}Usage: make [command]${NC}"
	@echo ""
	@echo "${YELLOW}🐳  Docker Commands:${NC}"
	@echo "${GREEN}  docker-build${NC}   - ${ROCKET}	Build and start Docker containers in detached mode"
	@echo "${GREEN}  docker-clean${NC}   - ${TRASH}	Clean up Docker resources (images, containers, volumes, caches)"
	@echo ""
	@echo "${YELLOW}📚  Documentation Commands:${NC}"
	@echo "${GREEN}  gen-docs${NC}       - ${DOCS}	Generate Swagger documentation"
	@echo ""
	@echo "${YELLOW}🛠️   Utility Commands:${NC}"
	@echo "${GREEN}  help${NC}           - ${INFO}	Show this help message"

# 🚀 Build and Start Docker Containers
.PHONY: docker-build
docker-build:
	@echo "${GREEN}${ROCKET} Building and starting Docker containers...${NC}"
	docker-compose up -d --build
	@echo "${GREEN}${CHECK} Docker containers are up and running!${NC}"

# 🗑️ Clean Docker Resources
.PHONY: docker-clean
docker-clean:
	@echo "${YELLOW}${TRASH} Cleaning up Docker resources...${NC}"
	@echo "${YELLOW}  - Stopping and removing containers...${NC}"
	docker-compose down --remove-orphans
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