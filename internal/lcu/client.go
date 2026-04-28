package lcu

import (
	"encoding/json"
	"fmt"

	lcu "github.com/its-haze/lcu-gopher"
)

// Client LCU 客户端封装
type Client struct {
	lcuClient *lcu.Client
	eventBus  *EventBus
}

// NewClient 创建 LCU 客户端（自动检测 LOL 进程）
func NewClient() (*Client, error) {
	config := lcu.DefaultConfig()

	lcuClient, err := lcu.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("create lcu client: %w", err)
	}

	return &Client{
		lcuClient: lcuClient,
		eventBus:  NewEventBus(),
	}, nil
}

// Connect 连接 LOL 客户端
func (c *Client) Connect() error {
	if err := c.lcuClient.Connect(); err != nil {
		return fmt.Errorf("connect to LCU: %w", err)
	}
	return nil
}

// Disconnect 断开连接
func (c *Client) Disconnect() error {
	return c.lcuClient.Disconnect()
}

// StartListening 启动事件监听
func (c *Client) StartListening() error {
	// 监听游戏阶段变化
	if err := c.lcuClient.SubscribeToGamePhase(func(phase lcu.GamePhase) {
		c.eventBus.Publish(EventGamePhaseChanged, phase)
	}); err != nil {
		return fmt.Errorf("subscribe game phase: %w", err)
	}

	// 监听选人会话变化
	if err := c.lcuClient.Subscribe("/lol-champ-select/v1/session", func(event *lcu.Event) {
		var session lcu.ChampSelectSession
		data, _ := json.Marshal(event.Data)
		if err := json.Unmarshal(data, &session); err != nil {
			return
		}
		c.eventBus.Publish(EventChampSelectUpdate, session)
	}, lcu.EventTypeCreate, lcu.EventTypeUpdate); err != nil {
		return fmt.Errorf("subscribe champ select: %w", err)
	}

	return nil
}

// GetCurrentSummoner 获取当前召唤师信息
func (c *Client) GetCurrentSummoner() (*lcu.Summoner, error) {
	return c.lcuClient.GetCurrentSummoner()
}

// GetChampSelectSession 获取当前选人会话
func (c *Client) GetChampSelectSession() (*lcu.ChampSelectSession, error) {
	session, err := c.lcuClient.GetChampSelectSession()
	if err != nil {
		return nil, fmt.Errorf("get champ select session: %w", err)
	}

	return session, nil
}

// GetGamePhase 获取当前游戏阶段
func (c *Client) GetGamePhase() (string, error) {
	resp, err := c.lcuClient.Get("/lol-gameflow/v1/gameflow-phase")
	if err != nil {
		return "", fmt.Errorf("get game phase: %w", err)
	}
	defer resp.Body.Close()

	var phase string
	if err := json.NewDecoder(resp.Body).Decode(&phase); err != nil {
		return "", fmt.Errorf("decode game phase: %w", err)
	}

	return phase, nil
}

// Subscribe 订阅自定义事件
func (c *Client) Subscribe(eventType EventType, handler EventHandler) {
	c.eventBus.Subscribe(eventType, handler)
}

// TeamMember 队友信息
type TeamMember struct {
	ChampionID int    `json:"champion_id"`
	CellID     int    `json:"cell_id"`
	SummonerID int64  `json:"summoner_id"`
	Position   string `json:"position"`
}

// GetMyTeam 获取我方队友列表
func (c *Client) GetMyTeam() ([]TeamMember, error) {
	session, err := c.GetChampSelectSession()
	if err != nil {
		return nil, err
	}

	var members []TeamMember
	for _, p := range session.MyTeam {
		members = append(members, TeamMember{
			ChampionID: p.ChampionId,
			CellID:     p.CellId,
			SummonerID: p.SummonerId,
			Position:   p.AssignedPosition,
		})
	}

	return members, nil
}

// GetEnemyTeam 获取敌方队伍列表
func (c *Client) GetEnemyTeam() ([]TeamMember, error) {
	session, err := c.GetChampSelectSession()
	if err != nil {
		return nil, err
	}

	var members []TeamMember
	for _, p := range session.TheirTeam {
		members = append(members, TeamMember{
			ChampionID: p.ChampionId,
			CellID:     p.CellId,
			SummonerID: p.SummonerId,
			Position:   p.AssignedPosition,
		})
	}

	return members, nil
}
