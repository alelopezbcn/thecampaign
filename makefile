mocks:
# board package
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/board/castle.go -destination ./test/mocks/castle_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/board/cemetery.go -destination ./test/mocks/cemetery_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/board/deck.go -destination ./test/mocks/deck_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/board/discardpile.go -destination ./test/mocks/discardpile_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/board/field.go -destination ./test/mocks/field_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/board/hand.go -destination ./test/mocks/hand_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/board/board.go -destination ./test/mocks/board_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/board/player.go -destination ./test/mocks/player_mocks.go
# cards package
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/ambush.go -destination ./test/mocks/ambush_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/attackablebase.go -destination ./test/mocks/attackable_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/bloodrain.go -destination ./test/mocks/bloodrain_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/cardbase.go -destination ./test/mocks/card_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/catapult.go -destination ./test/mocks/catapult_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/harpoon.go -destination ./test/mocks/harpoon_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/helper.go -destination ./test/mocks/dealer_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/fortress.go -destination ./test/mocks/fortress_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/resources.go -destination ./test/mocks/resource_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/resurrection.go -destination ./test/mocks/resurrection_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/specialpower.go -destination ./test/mocks/specialpower_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/spy.go -destination ./test/mocks/spy_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/thief.go -destination ./test/mocks/thief_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/sabotage.go -destination ./test/mocks/sabotage_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/treason.go -destination ./test/mocks/treason_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/warriors.go -destination ./test/mocks/warrior_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/cards/weapons.go -destination ./test/mocks/weapon_mocks.go
# gameactions package
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/gameactions/gameaction.go -destination ./test/mocks/gameaction_mocks.go
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package mocks -source ./internal/domain/gameactions/game_roles.go -destination ./test/mocks/gameroles_mocks.go
# websocket package
	cd backend && go run go.uber.org/mock/mockgen@v0.4.0 -package websocket -source ./internal/websocket/hub.go -destination ./internal/websocket/hubgame_mocks.go

up:
	docker-compose up --build

down:
	docker-compose down

logs:
	docker-compose logs -f

docker-tag-push: TAG = $(shell git describe --tags --abbrev=0)
docker-tag-push:
	@echo Building and pushing thecampaign:$(TAG)
	docker login
	docker build -t thecampaign:$(TAG) .
	docker tag thecampaign:$(TAG) alelopezcop/thecampaign:$(TAG)
	docker push alelopezcop/thecampaign:$(TAG)

test:
	cd backend && go test -mod=vendor -cover ./...

test-verbose:
	cd backend && go test -mod=vendor -cover -v ./...

tidy:
	cd backend && go mod tidy

fmt:
	cd backend && go fmt ./...