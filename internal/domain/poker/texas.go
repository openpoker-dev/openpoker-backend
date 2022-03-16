package poker

import (
	"errors"
	"fmt"
	"sync"

	"github.com/openpoker-dev/contrib/card"
)

type (
	// ActionSpecification 行动的抽象
	ActionSpecification interface {
		Action() ActionOption
		Description() string
		SuggestBettingSize() Chip
		CanPerform(betting Chip, remaining Chip) (Chip, error)
	}

	PlayerAttributes interface {
		ID() PlayerID        // 玩家ID
		Chips() Chip         // 剩余筹码
		Position() Position  // 玩家位置
		State() BettingState // 玩家状态
	}

	// Player 对玩家行为的抽象
	Player interface {
		PlayerAttributes
		TakePosition(Position)                                         // 分配玩家位置
		Activate(ActionID, chan<- ActionEvent, ...ActionSpecification) // 轮到玩家行动
		Bet(ActionID, ActionOption, Chip) error                        // 执行动作, check/call/raise/bet/fold
		CancelBetting(ActionID) error                                  // 取消行动，比如超时未行动等 TODO: 取消该方法，将context通过Activate传递进取，Player内部自行判断是否超时
		AddCard(card.Card) error                                       // 发牌
		ShowDown() []card.Card                                         // 亮牌
		WinPot(Chip)                                                   // 赢牌
	}

	ActionEvent interface {
		Player() PlayerAttributes
		ActionOption() ActionOption
		BettingAmount() Chip
	}

	TexasPlayer struct {
		id             PlayerID
		holeCards      []card.Card // 手牌
		position       Position
		chips          Chip
		state          BettingState
		currentAction  ActionID
		specifications map[ActionOption]ActionSpecification
		callback       chan<- ActionEvent
		mu             sync.RWMutex
	}

	playerAttributes struct {
		id       PlayerID
		chips    Chip
		state    BettingState
		position Position
	}

	actionEvent struct {
		playerAttributes PlayerAttributes
		option           ActionOption
		amount           Chip
	}
)

var (
	_ Player = (*TexasPlayer)(nil)

	_ PlayerAttributes = playerAttributes{}
	_ PlayerAttributes = (*TexasPlayer)(nil)

	_ ActionEvent = actionEvent{}

	actionStateMapping = map[ActionOption]BettingState{
		ActionCheck: StateChecked,
		ActionBet:   StateBetted,
		ActionCall:  StateCalled,
		ActionRaise: StateRaised,
		ActionFold:  StateFolded,
	}
)

func NewTexasPlayer(id PlayerID, postion Position, chips Chip) Player {
	return &TexasPlayer{
		id:             id,
		position:       postion,
		chips:          chips,
		holeCards:      make([]card.Card, 0, 2),
		state:          StateJoined,
		specifications: make(map[ActionOption]ActionSpecification),
		callback:       nil,
	}
}

func (p *TexasPlayer) ID() PlayerID {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.id
}

func (p *TexasPlayer) Chips() Chip {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.chips
}

func (p *TexasPlayer) Position() Position {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.position
}

func (p *TexasPlayer) TakePosition(pos Position) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.position = pos
}

func (p *TexasPlayer) State() BettingState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

func (p *TexasPlayer) isAllIn() {
	if p.chips == 0 && p.state != StateFolded {
		p.state = StateAllIn
	}
}

func (p *TexasPlayer) Activate(aid ActionID, callback chan<- ActionEvent, specs ...ActionSpecification) {
	if len(specs) == 0 {
		return
	}

	p.mu.Lock()
	p.currentAction = aid
	p.callback = callback
	for k := range p.specifications {
		delete(p.specifications, k)
	}
	for _, spec := range specs {
		p.specifications[spec.Action()] = spec
	}
	p.mu.Unlock()

	if len(specs) == 1 { // 自动执行
		p.Bet(aid, specs[0].Action(), specs[0].SuggestBettingSize())
	}
}

func (p *TexasPlayer) Bet(aid ActionID, at ActionOption, c Chip) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentAction != aid {
		return fmt.Errorf("unmatched action id, expected:%s; actural:%s", p.currentAction, aid)
	}

	opt, ok := p.specifications[at]
	if !ok {
		return errors.New("disallowed action type")
	}

	bettingChips, err := opt.CanPerform(c, p.chips)
	if err != nil {
		return err
	}

	defer func() {
		p.isAllIn()
		p.callback <- newActionEvent(p.cloneAttributes(), at, bettingChips)
	}()
	p.chips -= bettingChips

	if state, ok := actionStateMapping[at]; ok {
		p.state = state
	}
	return nil
}

func (p *TexasPlayer) CancelBetting(aid ActionID) error {
	p.mu.RLock()
	if p.currentAction != aid {
		p.mu.RUnlock()
		return fmt.Errorf("unmatched action id: %s", aid)
	}

	// 有check则check，否则直接盖牌
	cloned := make(map[ActionOption]ActionSpecification)
	for k, v := range p.specifications {
		cloned[k] = v
	}
	p.mu.RUnlock()

	var foldSpec ActionSpecification
	for opt, spec := range cloned {
		if opt == ActionCheck {
			return p.Bet(aid, opt, spec.SuggestBettingSize())
		}
		if opt == ActionFold && spec != nil {
			foldSpec = spec
		}
	}

	if foldSpec != nil {
		return p.Bet(aid, foldSpec.Action(), foldSpec.SuggestBettingSize())
	}
	return errors.New("unreachable code")
}

func (p *TexasPlayer) AddCard(card card.Card) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.holeCards) < 2 {
		p.holeCards = append(p.holeCards, card)
	}
	return errors.New("too many cards taken")
}

func (p *TexasPlayer) ShowDown() []card.Card {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.holeCards[:]
}

func (p *TexasPlayer) WinPot(c Chip) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.chips += c
}

func (p *TexasPlayer) cloneAttributes() PlayerAttributes {
	return playerAttributes{
		id:       p.id,
		state:    p.state,
		chips:    p.chips,
		position: p.position,
	}
}

func (ps playerAttributes) ID() PlayerID {
	return ps.id
}

func (ps playerAttributes) Position() Position {
	return ps.position
}

func (ps playerAttributes) Chips() Chip {
	return ps.chips
}

func (ps playerAttributes) State() BettingState {
	return ps.state
}

func newActionEvent(player PlayerAttributes, opt ActionOption, amount Chip) ActionEvent {
	return actionEvent{
		playerAttributes: player,
		option:           opt,
		amount:           amount,
	}
}

func (ae actionEvent) Player() PlayerAttributes {
	return ae.playerAttributes
}

func (ae actionEvent) ActionOption() ActionOption {
	return ae.option
}

func (ae actionEvent) BettingAmount() Chip {
	return ae.amount
}
