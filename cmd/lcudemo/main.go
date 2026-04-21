package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"haxPlugins/internal/lcu"
	"haxPlugins/internal/logger"
)

func main() {
	fmt.Println("=== LCU 客户端连接测试 ===")
	fmt.Println()

	if err := logger.Init(); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	// 创建 LCU 客户端
	client, err := lcu.NewClient()
	if err != nil {
		log.Fatalf("创建 LCU 客户端失败: %v", err)
	}

	// 连接 LOL 客户端
	fmt.Println("正在连接 LOL 客户端...")
	if err := client.Connect(); err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	fmt.Println("连接成功!")
	fmt.Println()

	// 获取当前召唤师信息
	fmt.Println("[1] 获取当前召唤师信息...")
	summoner, err := client.GetCurrentSummoner()
	if err != nil {
		log.Printf("  获取召唤师信息失败: %v", err)
	} else {
		fmt.Printf("  召唤师名称: %s\n", summoner.DisplayName)
		fmt.Printf("  等级: %d\n", summoner.SummonerLevel)
		fmt.Printf("  ID: %d\n", summoner.SummonerID)
	}
	fmt.Println()

	// 获取当前游戏阶段
	fmt.Println("[2] 获取当前游戏阶段...")
	phase, err := client.GetGamePhase()
	if err != nil {
		log.Printf("  获取游戏阶段失败: %v", err)
	} else {
		fmt.Printf("  当前阶段: %s\n", phase)
	}
	fmt.Println()

	// 获取选人会话
	fmt.Println("[3] 获取选人会话...")
	session, err := client.GetChampSelectSession()
	if err != nil {
		fmt.Printf("  未在选人阶段: %v\n", err)
	} else {
		fmt.Printf("  游戏 ID: %d\n", session.GameId)
		fmt.Printf("  我方人数: %d\n", len(session.MyTeam))
		fmt.Printf("  对方人数: %d\n", len(session.TheirTeam))
		for _, p := range session.MyTeam {
			fmt.Printf("    - 英雄: %d, 召唤师: %d\n", p.ChampionId, p.SummonerId)
		}
	}
	fmt.Println()

	// 启动事件监听
	fmt.Println("[4] 启动事件监听 (按 Ctrl+C 停止)...")
	fmt.Println()

	client.Subscribe(lcu.EventGamePhaseChanged, func(data interface{}) {
		fmt.Printf("[事件] 游戏阶段变化: %v\n", data)
	})

	client.Subscribe(lcu.EventChampSelectUpdate, func(data interface{}) {
		fmt.Printf("[事件] 选人会话更新\n")
	})

	if err := client.StartListening(); err != nil {
		log.Printf("启动事件监听失败: %v", err)
	}

	// 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 定期打印心跳
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			fmt.Println()
			fmt.Println("正在断开连接...")
			client.Disconnect()
			fmt.Println("已断开")
			return
		case <-ticker.C:
			phase, err := client.GetGamePhase()
			if err != nil {
				fmt.Printf("[心跳] 获取阶段失败: %v\n", err)
			} else {
				fmt.Printf("[心跳] 当前阶段: %s\n", phase)
			}
		}
	}
}
