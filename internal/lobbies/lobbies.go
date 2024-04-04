package lobbies

import (
	"net/http"
	"sync"
	"time"
)

type Lobbies struct {
	// mux routes the various endpoints to the appropriate handler.
	mux http.ServeMux

	lobbies map[string]*Lobby

	mu sync.Mutex
}

func NewLobbies(authenticate func(http.Handler) http.Handler) *Lobbies {
	l := &Lobbies{
		lobbies: make(map[string]*Lobby),
	}
	go l.clean()
	l.mux.Handle("/lobby/create", authenticate(http.HandlerFunc(l.createHandler)))
	l.mux.Handle("/lobby/connect", authenticate(http.HandlerFunc(l.connectHandler)))
	l.mux.HandleFunc("/health", healthHandler)
	return l
}

func (l *Lobbies) clean() {
	for range time.Tick(time.Minute * 30) {
		for _, lobby := range l.lobbies {
			if lobby.createdAt.Add(time.Minute * 15).After(time.Now()) {
				lobby.clean()
			}
		}
	}
}
