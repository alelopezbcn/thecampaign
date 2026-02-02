# Plan de Implementación: Sistema de Fases

## Resumen de Cambios en el Dominio (Ya Implementados)

Los métodos de `Game` ahora devuelven `(GameStatus, error)` y la acción actual se gestiona automáticamente:
- `DrawCard` → establece `ActionTypeAttack`
- `Attack`, `SpecialPower`, `Catapult` → establecen `ActionTypeSpySteal`
- `Spy`, `Steal` → establecen `ActionTypeBuy`
- `Buy` → establece `ActionTypeConstruct`
- `Construct` → establece `ActionTypeEndTurn`
- `EndTurn` → establece `ActionTypeDrawCard`

`GameStatus` ahora incluye:
- `CurrentAction string` - la fase actual
- `NewCards []string` - IDs de las cartas nuevas
- `CanSwap bool` - si puede hacer swap

`HandCard` usa `CanBeUsed` en lugar de `CanConstruct`.

---

## Tareas Backend

### 1. websocket/message.go
- Actualizar `GameStatusDTO`:
  - Agregar `CurrentAction string`
  - Agregar `NewCards []string`
  - Agregar `CanSwap bool`
  - Eliminar: `CanAttack`, `CanCatapult`, `CanSpy`, `CanSteal`, `CanBuy`, `CanInitiateCastle`, `CanGrowCastle`
- Actualizar `HandCardDTO`:
  - Cambiar `CanConstruct` por `CanBeUsed`

### 2. websocket/dto.go
- Actualizar `ConvertGameStatus`:
  - Mapear los nuevos campos `CurrentAction`, `NewCards`, `CanSwap`
  - Eliminar mapeo de campos `Can*` eliminados
- Actualizar `convertHandCards`:
  - Cambiar `CanConstruct` por `CanBeUsed`

### 3. websocket/hub.go
- Actualizar `executeGameAction`:
  - Cambiar firma para aceptar `func(*domain.Game) (gamestatus.GameStatus, error)`
  - Usar el `GameStatus` retornado para enviar el estado
- Actualizar `sendGameState`:
  - El método `NewGameStatus` ahora requiere `action types.ActionType`
  - Necesita obtener la acción actual del juego

### 4. websocket/handlers.go
- Actualizar todos los handlers para usar la nueva firma de métodos:
  - `handleAttack`: `g.Attack` devuelve `(GameStatus, error)`
  - `handleSpecialPower`: `g.SpecialPower` devuelve `(GameStatus, error)`
  - `handleMoveWarrior`: `g.MoveWarriorToField` devuelve `(GameStatus, error)`
  - `handleTrade`: `g.Trade` devuelve `(GameStatus, error)`
  - `handleBuy`: `g.Buy` devuelve `(GameStatus, error)`
  - `handleConstruct`: `g.Construct` devuelve `(GameStatus, error)`
  - `handleSteal`: `g.Steal` devuelve `(GameStatus, error)`
  - `handleCatapult`: `g.Catapult` devuelve `(GameStatus, error)`
- `handleSetInitialWarriors`:
  - Cambiar `DrawCards` por `DrawCard`
  - Manejar el `GameStatus` retornado
- `handleEndTurn`:
  - `EndTurn` devuelve `(GameStatus, error)`
  - `DrawCard` en lugar de `DrawCards`
- `handleSpy`:
  - `Spy` devuelve `([]ports.Card, GameStatus, error)`
- Agregar nuevo handler `handleSkipPhase` para saltar fases

### 5. cmd/main.go
- Actualizar todas las llamadas que usan los métodos modificados
- Eliminar referencias a `GetStatusForNextPlayer`
- Eliminar referencias a campos `Can*` eliminados
- Manejar los valores de retorno `(GameStatus, error)`

---

## Tareas Frontend

### 6. frontend/index.html
- Agregar indicador visual de fase actual
- Agregar botón "Skip Phase" / "Saltar Fase"

### 7. frontend/static/js/game.js
- Actualizar `handleGameState`:
  - Usar `current_action` para determinar la fase actual
- Actualizar `updateActionButtons`:
  - Habilitar/deshabilitar botones basándose en `current_action` en lugar de flags `can_*`
  - Usar `can_be_used` de las cartas en lugar de `can_construct`
- Agregar función `skipPhase` para saltar la fase actual
- Agregar renderizado del indicador de fase

### 8. frontend/static/css/styles.css
- Estilos para el indicador de fase
- Estilos para el botón de saltar fase

---

## Nuevos Mensajes WebSocket

### Mensaje para saltar fase
- Tipo: `skip_phase`
- Sin payload necesario
- El backend avanza a la siguiente fase automáticamente

---

## Orden de Implementación

1. Backend: message.go (DTOs)
2. Backend: dto.go (conversiones)
3. Backend: hub.go (executeGameAction y sendGameState)
4. Backend: handlers.go (todos los handlers)
5. Backend: cmd/main.go (CLI de pruebas)
6. Frontend: index.html (UI)
7. Frontend: game.js (lógica)
8. Frontend: styles.css (estilos)

---

## Mapeo de Fases a UI

| ActionType | Fase UI | Acciones Permitidas |
|------------|---------|---------------------|
| `draw` | Robar Carta | Solo Draw (automático) |
| `attack` | Atacar | Attack, SpecialPower, Catapult, Skip |
| `spy/steal` | Espiar/Robar | Spy, Steal, Skip (auto si no tiene cartas) |
| `buy` | Comprar | Buy, Skip |
| `construct` | Construir | Construct, Skip |
| `endturn` | Fin de Turno | Solo EndTurn |

**Nota**: MoveWarrior y Trade están disponibles en cualquier fase (siempre que sea tu turno).
