package lcu

import "sync"

// EventType 事件类型
type EventType string

const (
	// EventGamePhaseChanged 游戏阶段变化
	EventGamePhaseChanged EventType = "game_phase_changed"
	// EventChampSelectUpdate 选人会话更新
	EventChampSelectUpdate EventType = "champ_select_update"
	// EventGameStart 游戏开始
	EventGameStart EventType = "game_start"
	// EventEnterChampSelect 进入选人阶段
	EventEnterChampSelect EventType = "enter_champ_select"
)

// EventHandler 事件处理函数
type EventHandler func(data interface{})

// EventBus 事件总线
type EventBus struct {
	handlers map[EventType][]EventHandler
	mu       sync.RWMutex
}

// NewEventBus 创建事件总线
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]EventHandler),
	}
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Publish 发布事件
func (eb *EventBus) Publish(eventType EventType, data interface{}) {
	eb.mu.RLock()
	handlers := make([]EventHandler, len(eb.handlers[eventType]))
	copy(handlers, eb.handlers[eventType])
	eb.mu.RUnlock()

	for _, h := range handlers {
		go h(data)
	}
}
