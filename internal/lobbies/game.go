package lobbies

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	qgn "github.com/quibbble/quibbble-controller/pkg/gamenotation"
)

func CreateGame(url, key, id string, teams []string, players map[string][]string) error {

	// create the initial QGN
	snapshot := &qgn.Snapshot{
		Tags: map[string]string{
			qgn.KeyTag:   key,
			qgn.IDTag:    id,
			qgn.TeamsTag: strings.Join(teams, ", "),
		},
	}
	for team, p := range players {
		tag := fmt.Sprintf("%s_%s", team, qgn.PlayersTagSuffix)
		snapshot.Tags[tag] = strings.Join(p, ", ")
	}

	resp, err := http.Post(url, "application/qgn", bytes.NewBuffer([]byte(snapshot.String())))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed with error %s", http.StatusText(resp.StatusCode))
	}
	return nil
}
