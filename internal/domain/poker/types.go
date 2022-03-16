package poker

type (
	// GameRound 轮次
	GameRound int

	// Position 位置
	Position int

	// BettingState 玩家状态
	BettingState int

	// ActionOption 动作
	ActionOption int

	// ActionID 行动ID
	ActionID string

	// MandatoryBlindPayment 强制下注
	MandatoryBlindPayment int

	// PlayerID 玩家ID
	PlayerID string

	// Chip 筹码
	Chip = int32
)

const (
	RoundPreFlop GameRound = iota + 1 // 翻前
	RoundFlop                         // 翻牌
	RoundTurn                         // 转牌
	RoundRiver                        // 河牌
)

const (
	PositionSmallBlind  Position = iota + 1 // SB
	PositionBigBlind                        // BB
	PositionUnderTheGun                     // UTG
	PositionUTGPlusOne                      // UTG+1
	PositionUTGPlusTwo                      // UTG+2
	PositionLoJack                          // LJ
	PositionHijack                          // HJ
	PositionCutOff                          // CO
	PositionButton                          // Dealer/Button
	PositionStraddle                        // 抓位
)

const (
	ActionCheck        ActionOption = iota + 1 // 过牌
	ActionFold                                 // 盖牌
	ActionBet                                  // 下注
	ActionCall                                 // 跟注
	ActionRaise                                // 加注
	ActionMandatoryBet                         // 强制下注，如SB/BB/ANTE/STRADDLE
)

const (
	StateJoined  BettingState = iota // 默认状态
	StateChecked                     // 已过牌
	StateBetted                      // 主动下注
	StateCalled                      // 已跟注
	StateRaised                      // 主动加注
	StateAllIn                       // 全下
	StateFolded                      // 已盖牌
)
