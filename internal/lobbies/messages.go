package lobbies

import (
	"encoding/json"
)

func (l *Lobby) broadcastConnectionMessage() {
	undecided := []string{}
	usernames := make(map[string]string)
	for p := range l.connected {
		team := l.team(p.player.UserID)
		if team == nil {
			undecided = append(undecided, p.player.UserID)
		}
		usernames[p.player.UserID] = p.player.Username
	}
	payload, _ := json.Marshal(Message{
		Type: ConnectionType,
		Details: ConnectionDetails{
			Players:   l.players,
			Usernames: usernames,
			Teams:     l.teams,
			Undecided: undecided,
		},
	})
	for p := range l.connected {
		l.sendMessage(p, payload)
	}
}

func (l *Lobby) broadcastCreatedMessage() {
	payload, _ := json.Marshal(Message{
		Type: CreatedType,
		Details: CreatedDetails{
			Key: l.key,
			ID:  l.id,
		},
	})
	for p := range l.connected {
		l.sendMessage(p, payload)
	}
}

func (l *Lobby) sendErrorMessage(player *Connection, err error) {
	payload, _ := json.Marshal(Message{
		Type:    ErrorType,
		Details: err.Error(),
	})
	l.sendMessage(player, payload)
}

func (l *Lobby) sendMessage(player *Connection, payload []byte) {
	select {
	case player.outputCh <- payload:
	default:
		delete(l.connected, player)
		go player.Close()
	}
}
