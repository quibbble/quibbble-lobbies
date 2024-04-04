package lobbies

import (
	"net/http"
	"sync"
)

type Lobbies struct {
	// mux routes the various endpoints to the appropriate handler.
	mux http.ServeMux

	lobbies map[string]*Lobby

	mu sync.Mutex
}

func NewLobbies(authenticate func(http.Handler) http.Handler) *Lobbies {
	gs := &Lobbies{
		lobbies: make(map[string]*Lobby),
	}
	gs.mux.Handle("/lobby/create", authenticate(http.HandlerFunc(gs.createHandler)))
	gs.mux.Handle("/lobby/connect", authenticate(http.HandlerFunc(gs.connectHandler)))
	gs.mux.HandleFunc("/health", healthHandler)
	return gs
}
