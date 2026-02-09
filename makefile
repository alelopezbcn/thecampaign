mocks:
	go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./backend/internal/domain/ports/card.go -destination ./backend/test/mocks/card_mocks.go
	go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./backend/internal/domain/ports/castle.go -destination ./backend/test/mocks/castle_mocks.go
	go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./backend/internal/domain/ports/deck.go -destination ./backend/test/mocks/deck_mocks.go
	go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./backend/internal/domain/ports/field.go -destination ./backend/test/mocks/field_mocks.go
	go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./backend/internal/domain/ports/hand.go -destination ./backend/test/mocks/hand_mocks.go
	go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./backend/internal/domain/ports/observers.go -destination ./backend/test/mocks/observers_mocks.go
	go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./backend/internal/domain/ports/player.go -destination ./backend/test/mocks/player_mocks.go
	go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./backend/internal/domain/ports/dealer.go -destination ./backend/test/mocks/dealer_mocks.go
	go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./backend/internal/domain/ports/cemetery.go -destination ./backend/test/mocks/cemetery_mocks.go
	go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./backend/internal/domain/ports/discardpile.go -destination ./backend/test/mocks/discardpile_mocks.go
	go run go.uber.org/mock/mockgen@v0.4.0 -package domain -source ./backend/internal/domain/gamestatusprovider.go -destination ./backend/internal/domain/gamestatusprovider_mocks.go

up:
	docker-compose up --build

down:
	docker-compose down

logs:
	docker-compose logs -f

## example: make docker-tag-push TAG=yourtag
docker-tag-push:
	@if [ -z "$(TAG)" ]; then \
		TAG=$$(git describe --tags --abbrev=0); \
		if [ -z "$$TAG" ]; then \
			echo "TAG variable is required and no git tags found"; \
			exit 1; \
		fi; \
	else \
		TAG=$(TAG); \
	fi; \
	docker build -t thecampaign:$$TAG .; \
	docker tag thecampaign:$$TAG alelopezcop/thecampaign:$$TAG; \
	docker push alelopezcop/thecampaign:$$TAG
