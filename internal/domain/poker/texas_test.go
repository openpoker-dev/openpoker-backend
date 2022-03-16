package poker

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestPlayer(t *testing.T) {
	player := NewTexasPlayer(PlayerID(uuid.NewString()), PositionBigBlind, 1000)
	fmt.Println(player.Chips())
}
