package lobbies

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/quibbble/quibbble-controller/pkg/auth"
	"nhooyr.io/websocket"
)

func (l *Lobbies) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.mux.ServeHTTP(w, r)
}

func (l *Lobbies) createHandler(w http.ResponseWriter, r *http.Request) {

	type Create struct {
		Key   string `json:"key"`
		ID    string `json:"id"`
		Teams int    `json:"teams"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var create Create
	if err := json.Unmarshal(body, &create); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("%s_%s", create.Key, create.ID)
	_, ok := l.lobbies[key]
	if ok {
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	var colors = []string{
		"red", "blue", "green", "yellow", "orange", "pink", "purple", "teal",
	}

	if create.Teams > len(colors) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	clean := func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		delete(l.lobbies, create.ID)
	}

	lobby := NewLobby(create.Key, create.ID, colors[:create.Teams], clean)
	go lobby.Start()
	l.lobbies[key] = lobby

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(http.StatusText(http.StatusCreated)))
}

func (l *Lobbies) connectHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(auth.UserID).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	username, ok := r.Context().Value(auth.Username).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	key := r.PathValue("key")
	id := r.PathValue("id")
	lobby, ok := l.lobbies[fmt.Sprintf("%s_%s", key, id)]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // allow origin checks are handled at the ingress
	})
	if err != nil {
		log.Println(err.Error())
		return
	}

	p := NewConnection(&Player{userId, username}, conn, lobby.inputCh)
	lobby.joinCh <- p

	ctx := context.Background()
	go func() {
		if err := p.ReadPump(ctx); err != nil {
			log.Println(err.Error())
		}
		lobby.leaveCh <- p
		p.conn.CloseNow()
	}()
	go func() {
		if err := p.WritePump(ctx); err != nil {
			log.Println(err.Error())
		}
	}()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}
