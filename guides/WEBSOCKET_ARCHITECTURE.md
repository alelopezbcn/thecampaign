# Arquitectura WebSocket - The Campaign

## Índice
1. [¿Qué es WebSocket?](#qué-es-websocket)
2. [Arquitectura General](#arquitectura-general)
3. [Estructura de Archivos](#estructura-de-archivos)
4. [Flujo Completo de una Conexión](#flujo-completo-de-una-conexión)
5. [Componentes del Backend](#componentes-del-backend)
6. [Componentes del Frontend](#componentes-del-frontend)
7. [Flujo de Mensajes](#flujo-de-mensajes)
8. [Ejemplos de Flujos Específicos](#ejemplos-de-flujos-específicos)
9. [Guía de Debugging](#guía-de-debugging)

---

## ¿Qué es WebSocket?

### HTTP vs WebSocket

**HTTP tradicional:**
```
Cliente ──request──> Servidor
Cliente <──response── Servidor
(conexión se cierra)
```
- El cliente siempre inicia la comunicación
- Cada request abre una nueva conexión
- El servidor NO puede enviar datos sin que el cliente pregunte primero

**WebSocket:**
```
Cliente ══════════════════════ Servidor
        <── mensajes en ambas ──>
           direcciones en tiempo
              real, conexión
               permanente
```
- Conexión persistente (se abre una vez y queda abierta)
- Bidireccional: ambos pueden enviar mensajes en cualquier momento
- Ideal para juegos, chats, aplicaciones en tiempo real

### ¿Por qué WebSocket para este juego?
- Cuando un jugador hace una acción, el otro debe verla inmediatamente
- El servidor necesita notificar a ambos jugadores de cambios
- Sin WebSocket, tendrías que hacer "polling" (preguntar constantemente al servidor)

---

## Arquitectura General

```
┌─────────────────────────────────────────────────────────────────────┐
│                           SERVIDOR GO                                │
│                                                                      │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────────────┐  │
│  │   server.go │──────│    Hub      │──────│    GameRoom         │  │
│  │ (HTTP+WS)   │      │ (gestiona   │      │ (una sala por       │  │
│  │             │      │  conexiones)│      │  partida)           │  │
│  └─────────────┘      └─────────────┘      │                     │  │
│         │                    │              │  - Game (lógica)    │  │
│         │                    │              │  - Player1 Client   │  │
│         │                    │              │  - Player2 Client   │  │
│         ▼                    │              └─────────────────────┘  │
│  ┌─────────────┐             │                                       │
│  │   Client    │◄────────────┘                                       │
│  │ (1 por cada │                                                     │
│  │  conexión)  │                                                     │
│  └─────────────┘                                                     │
└─────────────────────────────────────────────────────────────────────┘
         ▲                              ▲
         │ WebSocket                    │ WebSocket
         │ conexión                     │ conexión
         ▼                              ▼
┌─────────────────┐            ┌─────────────────┐
│   BROWSER 1     │            │   BROWSER 2     │
│   (Jugador 1)   │            │   (Jugador 2)   │
│                 │            │                 │
│   game.js       │            │   game.js       │
└─────────────────┘            └─────────────────┘
```

---

## Estructura de Archivos

```
backend/
├── cmd/server/
│   └── main.go              # Punto de entrada, inicia el servidor
│
├── internal/
│   ├── server/
│   │   └── server.go        # Configura HTTP y rutas
│   │
│   ├── websocket/
│   │   ├── hub.go           # Centro de control de conexiones y salas
│   │   ├── client.go        # Representa una conexión WebSocket individual
│   │   ├── handlers.go      # Procesa acciones del juego (attack, draw, etc)
│   │   ├── message.go       # Define tipos de mensajes JSON
│   │   └── dto.go           # Convierte objetos del juego a JSON
│   │
│   └── domain/
│       └── game.go          # Lógica del juego (ya existía)

frontend/
├── index.html               # Interfaz del juego
└── static/
    ├── css/styles.css       # Estilos
    └── js/game.js           # Cliente WebSocket + UI
```

---

## Flujo Completo de una Conexión

### Paso 1: El navegador carga la página

```
Browser                          Servidor
   │                                │
   │──── GET /  ───────────────────>│  (HTTP normal)
   │<─── index.html ────────────────│
   │                                │
   │──── GET /static/js/game.js ───>│
   │<─── game.js ───────────────────│
```

**Archivo:** `server.go` líneas 25-28
```go
// Sirve archivos estáticos (HTML, CSS, JS)
mux.Handle("/", http.FileServer(http.Dir("../frontend")))
```

### Paso 2: El usuario hace clic en "Join Game"

En `game.js`, la función `joinGame()` se ejecuta:

```javascript
// game.js línea 196-216
function joinGame() {
    const playerName = document.getElementById('player-name').value;
    const gameID = document.getElementById('game-id').value;

    gameState.playerName = playerName;
    gameState.gameID = gameID;

    connectWebSocket();  // <-- Aquí se abre el WebSocket
}
```

### Paso 3: Se abre la conexión WebSocket

```javascript
// game.js línea 56-85
function connectWebSocket() {
    // Crea la conexión WebSocket al servidor
    ws = new WebSocket(`ws://${window.location.host}/ws`);

    ws.onopen = () => {
        // Conexión establecida
    };

    ws.onmessage = (event) => {
        // Cuando llega un mensaje del servidor
        const msg = JSON.parse(event.data);
        handleMessage(msg);
    };
}
```

**En el servidor**, cuando llega una petición a `/ws`:

```
Browser                          Servidor
   │                                │
   │──── GET /ws ──────────────────>│
   │     Upgrade: websocket         │
   │                                │
   │<─── 101 Switching Protocols ───│
   │                                │
   │════ CONEXIÓN WEBSOCKET ════════│
   │     (ahora es bidireccional)   │
```

**Archivo:** `server.go` línea 30
```go
mux.HandleFunc("/ws", s.hub.HandleWebSocket)
```

**Archivo:** `hub.go` líneas 50-70 (HandleWebSocket)
```go
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    // 1. "Upgradea" la conexión HTTP a WebSocket
    conn, err := upgrader.Upgrade(w, r, nil)

    // 2. Crea un objeto Client para esta conexión
    client := NewClient(h, conn)

    // 3. Registra el cliente en el Hub
    h.register <- client

    // 4. Inicia goroutines para leer/escribir mensajes
    go client.writePump()  // Envía mensajes al browser
    go client.readPump()   // Recibe mensajes del browser
}
```

### Paso 4: El cliente envía "join_game"

```javascript
// game.js - después de conectar
sendMessage('join_game', {
    player_name: 'Juan',
    game_id: 'partida123'
});

function sendMessage(type, payload) {
    ws.send(JSON.stringify({ type, payload }));
}
```

**Mensaje JSON enviado:**
```json
{
    "type": "join_game",
    "payload": {
        "player_name": "Juan",
        "game_id": "partida123"
    }
}
```

### Paso 5: El servidor recibe y procesa el mensaje

```
┌─────────────────────────────────────────────────────────────────┐
│                      FLUJO EN EL SERVIDOR                        │
│                                                                  │
│  client.go          hub.go              handlers.go              │
│  readPump()         processMessage()    handleJoinGame()         │
│      │                   │                    │                  │
│      │ 1. Lee JSON       │                    │                  │
│      │ del WebSocket     │                    │                  │
│      │                   │                    │                  │
│      │ 2. Deserializa    │                    │                  │
│      │ a Message{}       │                    │                  │
│      │                   │                    │                  │
│      ├──────────────────>│ 3. Switch por     │                  │
│      │ h.handleMessage   │ msg.Type          │                  │
│      │                   │                    │                  │
│      │                   ├───────────────────>│ 4. Ejecuta      │
│      │                   │ handleJoinGame()   │ la lógica       │
│      │                   │                    │                  │
│      │                   │                    │ 5. Crea/une     │
│      │                   │                    │ a GameRoom      │
│      │                   │                    │                  │
│      │                   │<───────────────────│ 6. Llama        │
│      │                   │                    │ sendGameState() │
│      │                   │                    │                  │
└─────────────────────────────────────────────────────────────────┘
```

**Archivo:** `client.go` líneas 70-100 (readPump)
```go
func (c *Client) readPump() {
    for {
        // Lee el mensaje del WebSocket
        _, message, err := c.conn.ReadMessage()

        // Deserializa JSON a struct Message
        var msg Message
        json.Unmarshal(message, &msg)

        // Envía al Hub para procesar
        c.hub.handleMessage <- ClientMessage{client: c, message: &msg}
    }
}
```

**Archivo:** `hub.go` líneas 91-123 (processMessage)
```go
func (h *Hub) processMessage(client *Client, msg *Message) {
    log.Printf("Processing message type: %s from player: %s",
               msg.Type, client.PlayerName)

    switch msg.Type {
    case MsgJoinGame:
        h.handleJoinGame(client, msg.Payload)
    case MsgSetInitialWarriors:
        h.handleSetInitialWarriors(client, msg.Payload)
    case MsgAttack:
        h.handleAttack(client, msg.Payload)
    case MsgEndTurn:
        h.handleEndTurn(client)
    // ... más casos
    }
}
```

---

## Componentes del Backend

### 1. Hub (`hub.go`)

El **Hub** es el "centro de control". Tiene estas responsabilidades:

```go
type Hub struct {
    clients       map[*Client]bool           // Todas las conexiones activas
    gameRooms     map[string]*GameRoom       // Salas de juego por ID
    register      chan *Client               // Canal para nuevas conexiones
    unregister    chan *Client               // Canal para desconexiones
    handleMessage chan ClientMessage         // Canal para mensajes entrantes
}
```

**Funciones principales:**

| Función | Qué hace |
|---------|----------|
| `Run()` | Loop infinito que procesa eventos (conexiones, desconexiones, mensajes) |
| `HandleWebSocket()` | Punto de entrada cuando un browser se conecta |
| `processMessage()` | Router que decide qué handler llamar según el tipo de mensaje |
| `handleJoinGame()` | Crea o une a una sala de juego |
| `sendGameState()` | Envía el estado del juego a todos los jugadores |

### 2. Client (`client.go`)

Representa **una conexión WebSocket individual** (un jugador):

```go
type Client struct {
    hub        *Hub              // Referencia al Hub
    conn       *websocket.Conn   // La conexión WebSocket real
    send       chan []byte       // Canal para mensajes salientes
    PlayerName string            // Nombre del jugador
    GameID     string            // ID de la partida
}
```

**Funciones principales:**

| Función | Qué hace |
|---------|----------|
| `readPump()` | Goroutine que lee mensajes del browser continuamente |
| `writePump()` | Goroutine que envía mensajes al browser continuamente |
| `SendMessage()` | Envía un mensaje específico al browser |
| `SendError()` | Envía un mensaje de error al browser |

**¿Por qué dos goroutines (readPump y writePump)?**

```
                    ┌──────────────┐
   Browser ────────>│  readPump()  │ Lee mensajes del browser
                    │  (goroutine) │ y los pasa al Hub
                    └──────────────┘

                    ┌──────────────┐
   Browser <────────│  writePump() │ Toma mensajes del canal 'send'
                    │  (goroutine) │ y los envía al browser
                    └──────────────┘
```

Esto permite que leer y escribir sean **independientes** y no se bloqueen mutuamente.

### 3. GameRoom (`hub.go`)

Una sala contiene todo lo necesario para una partida:

```go
type GameRoom struct {
    ID      string                    // "partida123"
    Game    *domain.Game              // La lógica del juego
    Players map[string]*Client        // "Juan" -> Client, "Pedro" -> Client
    mutex   sync.RWMutex              // Protege acceso concurrente
}
```

### 4. Handlers (`handlers.go`)

Cada acción del juego tiene su handler:

| Handler | Mensaje | Qué hace |
|---------|---------|----------|
| `handleSetInitialWarriors()` | `set_initial_warriors` | Coloca guerreros iniciales |
| `handleAttack()` | `attack` | Ejecuta un ataque |
| `handleEndTurn()` | `end_turn` | Termina el turno |
| `handleMoveWarrior()` | `move_warrior` | Mueve guerrero al campo |
| ... | ... | ... |

**Patrón común de un handler:**

```go
func (h *Hub) handleAttack(client *Client, payload interface{}) {
    // 1. Deserializar el payload
    data, _ := json.Marshal(payload)
    var p AttackPayload
    json.Unmarshal(data, &p)

    // 2. Obtener la sala de juego
    room, exists := h.getGameRoom(client)

    // 3. Ejecutar la lógica del juego (con lock para thread safety)
    room.mutex.Lock()
    err := room.Game.Attack(client.PlayerName, p.WarriorID, p.TargetID, p.WeaponID)
    room.mutex.Unlock()

    // 4. Enviar el nuevo estado a todos los jugadores
    h.sendGameState(client.GameID)
}
```

### 5. Messages (`message.go`)

Define la estructura de todos los mensajes:

```go
// Mensaje base
type Message struct {
    Type    MessageType `json:"type"`      // "join_game", "attack", etc.
    Payload interface{} `json:"payload"`   // Datos específicos
}

// Tipos de mensaje
const (
    MsgJoinGame         = "join_game"
    MsgSetInitialWarriors = "set_initial_warriors"
    MsgAttack           = "attack"
    MsgEndTurn          = "end_turn"
    MsgGameState        = "game_state"    // Servidor -> Cliente
    MsgError            = "error"         // Servidor -> Cliente
    // ...
)

// Payload para atacar
type AttackPayload struct {
    WarriorID string `json:"warrior_id"`
    TargetID  string `json:"target_id"`
    WeaponID  string `json:"weapon_id"`
}

// Payload del estado del juego (servidor -> cliente)
type GameStatePayload struct {
    GameStatus     GameStatusDTO `json:"game_status"`
    IsYourTurn     bool          `json:"is_your_turn"`
    GameEnded      bool          `json:"game_ended"`
    NewlyDrawnCard string        `json:"newly_drawn_card,omitempty"`
}
```

---

## Componentes del Frontend

### game.js - Estructura Principal

```javascript
// Estado global del juego
const gameState = {
    ws: null,              // Conexión WebSocket
    playerName: '',        // Nombre del jugador
    gameID: '',            // ID de la partida
    isYourTurn: false,     // ¿Es mi turno?
    currentState: null,    // Estado actual del juego
    selectedCards: [],     // Cartas seleccionadas
    newlyDrawnCard: null   // Carta recién robada (para animación)
};

// Conexión WebSocket
let ws = null;
```

### Funciones Principales

| Función | Qué hace |
|---------|----------|
| `connectWebSocket()` | Abre la conexión y configura callbacks |
| `sendMessage(type, payload)` | Envía un mensaje al servidor |
| `handleMessage(msg)` | Router para mensajes entrantes |
| `handleGameState(payload)` | Actualiza la UI con el nuevo estado |
| `renderGameBoard(status)` | Dibuja el tablero completo |
| `createCardElement(card)` | Crea el HTML de una carta |

### Flujo de un Mensaje Entrante

```javascript
ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    console.log('Received message:', msg);
    handleMessage(msg);
};

function handleMessage(msg) {
    switch (msg.type) {
        case 'game_state':
            handleGameState(msg.payload);
            break;
        case 'error':
            handleError(msg.payload);
            break;
        case 'game_started':
            handleGameStarted(msg.payload);
            break;
        // ...
    }
}

function handleGameState(payload) {
    // Actualiza el estado local
    gameState.isYourTurn = payload.is_your_turn;
    gameState.currentState = payload.game_status;
    gameState.newlyDrawnCard = payload.newly_drawn_card;

    // Decide qué pantalla mostrar
    if (!isSetupComplete(payload.game_status)) {
        showSetupScreen(payload.game_status);
    } else {
        showGameScreen(payload.game_status);
    }
}
```

---

## Flujo de Mensajes

### Diagrama de Secuencia: Unirse a una Partida

```
 Browser 1              Servidor                Browser 2
     │                     │                        │
     │ join_game           │                        │
     │ {player:"Juan",     │                        │
     │  game_id:"abc"}     │                        │
     │────────────────────>│                        │
     │                     │                        │
     │                     │ (crea GameRoom "abc")  │
     │                     │ (agrega Juan)          │
     │                     │                        │
     │   waiting_for_player│                        │
     │<────────────────────│                        │
     │                     │                        │
     │                     │         join_game      │
     │                     │     {player:"Pedro",   │
     │                     │      game_id:"abc"}    │
     │                     │<───────────────────────│
     │                     │                        │
     │                     │ (agrega Pedro a room)  │
     │                     │ (crea Game)            │
     │                     │                        │
     │    game_started     │      game_started      │
     │<────────────────────│───────────────────────>│
     │                     │                        │
     │    game_state       │      game_state        │
     │<────────────────────│───────────────────────>│
```

### Diagrama de Secuencia: Atacar

```
 Browser 1              Servidor                Browser 2
 (Juan - su turno)         │                    (Pedro)
     │                     │                        │
     │ attack              │                        │
     │ {warrior:"W1",      │                        │
     │  target:"W5",       │                        │
     │  weapon:"Sword1"}   │                        │
     │────────────────────>│                        │
     │                     │                        │
     │                     │ (valida que es turno   │
     │                     │  de Juan)              │
     │                     │                        │
     │                     │ (ejecuta               │
     │                     │  game.Attack())        │
     │                     │                        │
     │                     │ (calcula daño,         │
     │                     │  actualiza HP, etc)    │
     │                     │                        │
     │    game_state       │      game_state        │
     │ (perspectiva Juan)  │  (perspectiva Pedro)   │
     │<────────────────────│───────────────────────>│
     │                     │                        │
     │ (actualiza UI)      │      (actualiza UI)    │
```

### Perspectivas Diferentes

Cuando se envía `game_state`, cada jugador recibe información diferente:

```go
// hub.go - sendGameState()
for playerName, client := range room.Players {
    if playerName == currentPlayerName {
        // Jugador actual: ve su mano completa
        status = room.Game.GetStatusForNextPlayer()
    } else {
        // Oponente: ve su propia mano, no la del jugador actual
        status = domain.NewGameStatus(enemyPlayer, currentPlayer)
    }

    payload := GameStatePayload{
        GameStatus: ConvertGameStatus(status),
        IsYourTurn: playerName == currentPlayerName,
        // ...
    }

    client.SendMessage(MsgGameState, payload)
}
```

---

## Ejemplos de Flujos Específicos

### Flujo: Terminar Turno con Auto-Draw

Este es el flujo que implementamos para robar carta automáticamente:

```
 Browser 1              Servidor                Browser 2
 (Juan - su turno)         │                    (Pedro)
     │                     │                        │
     │ end_turn            │                        │
     │────────────────────>│                        │
     │                     │                        │
     │              ┌──────┴──────┐                 │
     │              │ handleEndTurn()               │
     │              │                               │
     │              │ 1. game.EndTurn("Juan")       │
     │              │    -> cambia turno a Pedro    │
     │              │                               │
     │              │ 2. Guardar mano de Pedro      │
     │              │    (antes de robar)           │
     │              │                               │
     │              │ 3. game.DrawCards("Pedro", 1) │
     │              │    -> Pedro roba 1 carta      │
     │              │                               │
     │              │ 4. Detectar carta nueva       │
     │              │    comparando manos           │
     │              │                               │
     │              │ 5. sendGameState(newCardID)   │
     │              └──────┬──────┘                 │
     │                     │                        │
     │    game_state       │      game_state        │
     │ {is_your_turn:false}│ {is_your_turn:true,    │
     │                     │  newly_drawn_card:"X"} │
     │<────────────────────│───────────────────────>│
     │                     │                        │
     │              (Pedro ve la carta              │
     │               nueva con animación)           │
```

**Código relevante en `handlers.go`:**

```go
func (h *Hub) handleEndTurn(client *Client) {
    // 1. Termina el turno actual
    room.Game.EndTurn(client.PlayerName)

    // 2. Obtiene el nuevo jugador actual
    nextPlayer, _ := room.Game.WhoIsCurrent()

    // 3. Guarda las cartas que tiene ANTES de robar
    handBefore := make(map[string]bool)
    for _, card := range nextPlayer.Hand().ShowCards() {
        handBefore[card.GetID()] = true
    }

    // 4. Roba una carta
    room.Game.DrawCards(nextPlayer.Name(), 1)

    // 5. Encuentra la carta nueva (la que no estaba antes)
    var newCardID string
    for _, card := range nextPlayer.Hand().ShowCards() {
        if !handBefore[card.GetID()] {
            newCardID = card.GetID()
            break
        }
    }

    // 6. Envía el estado con el ID de la carta nueva
    h.sendGameState(client.GameID, newCardID)
}
```

---

## Guía de Debugging

### 1. Logs del Servidor

Los logs actuales muestran:

```
Processing message type: join_game from player: Juan
Processing message type: set_initial_warriors from player: Juan
SetInitialWarriors: currentPlayer=Pedro, currentPlayerField=2, enemyField=1, bothHaveWarriors=true
Setup complete! Drawing card for Pedro (hand size before: 7)
Newly drawn card ID: G5 (hand size after: 8)
Sending game state with newCardID: G5
sendGameState to Pedro: isYourTurn=true, newlyDrawnCard='G5', handSize=8
sendGameState to Juan: isYourTurn=false, newlyDrawnCard='', handSize=7
```

**Qué buscar:**
- `Processing message type: X` → El servidor recibió el mensaje
- `SetInitialWarriors: bothHaveWarriors=true/false` → ¿Ambos pusieron guerreros?
- `Newly drawn card ID: X` → ¿Se identificó la carta robada?
- `sendGameState to X: newlyDrawnCard='Y'` → ¿Se envía el ID al cliente correcto?

### 2. Logs del Frontend (Consola del Browser)

```javascript
// Estos logs ya están en game.js
console.log('Game state updated:', payload);
console.log('Newly drawn card from payload:', payload.newly_drawn_card);
console.log('Creating card:', cardId, 'newlyDrawnCard:', gameState.newlyDrawnCard);
```

**Qué buscar:**
- `Newly drawn card from payload: X` → ¿El frontend recibe el ID?
- `Creating card: X newlyDrawnCard: X` → ¿Coinciden para aplicar animación?

### 3. Puntos de Debugging Clave

| Dónde | Qué verificar |
|-------|---------------|
| `hub.go:processMessage` | ¿Llega el mensaje correcto? |
| `handlers.go:handleEndTurn` | ¿Se ejecuta la lógica? |
| `hub.go:sendGameState` | ¿Se envía el newCardID? |
| `game.js:handleGameState` | ¿Se recibe newly_drawn_card? |
| `game.js:createCardElement` | ¿Se aplica la clase newly-drawn? |

### 4. Herramientas de Debugging

**Browser DevTools:**
- Pestaña "Network" → filtrar por "WS" → ver mensajes WebSocket
- Pestaña "Console" → ver logs de JavaScript

**Servidor:**
- Los logs aparecen en la terminal donde corre el servidor
- Añadir más `log.Printf()` donde necesites

### 5. Problemas Comunes

| Problema | Causa Probable | Solución |
|----------|----------------|----------|
| No aparecen logs nuevos | Servidor no recompilado | Ejecutar `go build` y reiniciar |
| `newly_drawn_card: undefined` | El servidor no envía el campo | Verificar que `newCardID` no esté vacío |
| Carta no se anima | ID no coincide | Comparar IDs en logs del servidor y browser |
| "Not your turn" error | Lógica de turnos | Verificar quién es `currentPlayer` |

---

## Resumen Visual

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           FLUJO COMPLETO                                 │
│                                                                          │
│  BROWSER                    SERVIDOR                       BROWSER       │
│  (Juan)                                                    (Pedro)       │
│    │                                                          │          │
│    │ 1. Click "Attack"                                        │          │
│    │    ↓                                                     │          │
│    │ 2. sendMessage('attack', {...})                          │          │
│    │    ↓                                                     │          │
│    │ 3. ws.send(JSON)                                         │          │
│    │    ════════════════════════>                             │          │
│    │                          4. client.readPump()            │          │
│    │                             ↓                            │          │
│    │                          5. hub.processMessage()         │          │
│    │                             ↓                            │          │
│    │                          6. handleAttack()               │          │
│    │                             ↓                            │          │
│    │                          7. game.Attack()                │          │
│    │                             ↓                            │          │
│    │                          8. sendGameState()              │          │
│    │                             ↓                            │          │
│    │    <════════════════════════════════════════════>        │          │
│    │ 9. ws.onmessage()                    9. ws.onmessage()   │          │
│    │    ↓                                    ↓                │          │
│    │ 10. handleGameState()              10. handleGameState() │          │
│    │    ↓                                    ↓                │          │
│    │ 11. renderGameBoard()              11. renderGameBoard() │          │
│    │                                                          │          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Próximos Pasos para Debugging del Auto-Draw

El problema actual es que `newly_drawn_card` llega como `undefined` al cliente. Para debuggear:

1. **Verifica los logs del servidor** - ¿Aparece `Newly drawn card ID: X`?
2. **Si no aparece** → El código de handlers.go no se está ejecutando (rebuild necesario)
3. **Si aparece** → Verifica `Payload JSON: {...}` para ver qué se envía
4. **En el browser** → Verifica que el payload tenga el campo `newly_drawn_card`

El documento te da el contexto completo para entender dónde puede estar fallando.
