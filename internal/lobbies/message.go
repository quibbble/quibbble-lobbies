package lobbies

// outgoing message types
const (
	ConnectionType = "connection"
	CreatedType    = "created"
	ErrorType      = "error"
)

// incoming message types
const (
	JoinType   = "join"
	CreateType = "create"
)

type Player struct {
	UserID   string `json:"uid"`
	Username string `json:"username"`
}

type Message struct {
	Type    string      `json:"type"`
	Details interface{} `json:"details"`
}

type Action struct {
	*Message
	*Connection
}

// outgoing messages
type ConnectionDetails struct {
	Players   map[string][]string `json:"players"`
	Usernames map[string]string   `json:"usernames"`
	Teams     []string            `json:"teams"`
	Undecided []string            `json:"undecided"`
}

type CreatedDetails struct {
	Key string `json:"key"`
	ID  string `json:"id"`
}

// incoming messages
type JoinDetails struct {
	Team string `json:"team"`
}
