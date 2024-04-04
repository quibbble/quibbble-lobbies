package lobbies

import (
	"fmt"
	"log"
	"runtime/debug"
	"slices"
	"time"

	"github.com/mitchellh/mapstructure"
)

type Lobby struct {
	// connected represents all players currently connected to the server.
	connected map[*Connection]struct{}

	// joinCh and leaveCh adds/remove a player from the server.
	joinCh, leaveCh chan *Connection

	// inputCh sends actions to the server to be processed.
	inputCh chan *Action

	// players is a map from team to list of players of the team.
	players map[string][]string

	// saved is a map from player id to team.
	saved map[string]*string

	// teams is the list of teams that needs to be filled.
	teams []string

	// key of the game to create.
	key string

	// id of the lobby and corresponding game once created.
	id string

	// clean function used to cleanup the lobby.
	clean func()

	// createdAt is when the lobby was created
	createdAt time.Time
}

func NewLobby(key, id string, teams []string, clean func()) *Lobby {
	l := &Lobby{
		connected: make(map[*Connection]struct{}),
		joinCh:    make(chan *Connection),
		leaveCh:   make(chan *Connection),
		inputCh:   make(chan *Action),
		players:   make(map[string][]string),
		teams:     teams,
		key:       key,
		id:        id,
		createdAt: time.Now(),
	}
	l.clean = func() {
		for connected := range l.connected {
			go connected.Close()
		}
		clean()
	}
	return l
}

func (l *Lobby) Start() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(string(debug.Stack()))
		}
	}()

	for {
		select {
		case p := <-l.joinCh:
			l.connected[p] = struct{}{}
			if team := l.saved[p.player.UserID]; team != nil {
				l.joinTeam(p.player.UserID, *team)
			}
			l.broadcastConnectionMessage()
		case p := <-l.leaveCh:
			team := l.leaveTeam(p.player.UserID)
			l.saved[p.player.UserID] = team
			delete(l.connected, p)
			go p.Close()
			l.broadcastConnectionMessage()
		case msg := <-l.inputCh:
			switch msg.Type {
			case JoinType:
				var details JoinDetails
				if err := mapstructure.Decode(msg.Details, &details); err != nil {
					l.sendErrorMessage(msg.Connection, err)
					continue
				}
				if !slices.Contains(l.teams, details.Team) {
					l.sendErrorMessage(msg.Connection, fmt.Errorf("provided team is invalid"))
					continue
				}
				l.joinTeam(msg.player.UserID, details.Team)
				l.broadcastConnectionMessage()
			case CreateType:
				ready := true
				if len(l.players) != len(l.teams) {
					ready = false
				}
				for _, players := range l.players {
					if len(players) < 1 {
						ready = false
						break
					}
				}
				if !ready {
					l.sendErrorMessage(msg.Connection, fmt.Errorf("not enough players have joined"))
					continue
				}
				url := "http://quibbble-controller.quibbble/create"
				if err := CreateGame(url, l.key, l.id, l.teams, l.players); err != nil {
					l.sendErrorMessage(msg.Connection, err)
					continue
				}
				l.broadcastCreatedMessage()
				l.clean()
				return
			}
		}
	}
}

func (l *Lobby) team(uid string) *string {
	var team *string
	for t, players := range l.players {
		if slices.Contains(players, uid) {
			team = &t
			break
		}
	}
	return team
}

func (l *Lobby) leaveTeam(uid string) *string {
	team := l.team(uid)
	if team != nil {
		l.players[*team] = slices.DeleteFunc(l.players[*team], func(it string) bool { return it == uid })
	}
	return team
}

func (l *Lobby) joinTeam(uid, team string) {
	l.leaveTeam(uid)
	l.players[team] = append(l.players[team], uid)
}
