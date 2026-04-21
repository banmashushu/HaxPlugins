import {useEffect, useState} from 'react';
import './App.css';
import {EventsOn} from "../wailsjs/runtime";
import {GetCurrentPhase} from "../wailsjs/go/main/App";
import ChampSelectView from "./components/ChampSelectView";

type GamePhase = 'None' | 'Lobby' | 'Matchmaking' | 'CheckedIntoTournament' |
    'ReadyCheck' | 'ChampSelect' | 'GameStart' | 'InProgress' |
    'WaitingForStats' | 'PreEndOfGame' | 'EndOfGame' | '';

function phaseDisplayName(phase: GamePhase): string {
    const names: Record<string, string> = {
        'None': '未连接',
        'Lobby': '大厅',
        'Matchmaking': '匹配中',
        'CheckedIntoTournament': '锦标赛',
        'ReadyCheck': '接受对局',
        'ChampSelect': '选择英雄',
        'GameStart': '游戏开始',
        'InProgress': '游戏进行中',
        'WaitingForStats': '等待结算',
        'PreEndOfGame': '游戏结束',
        'EndOfGame': '已结束',
    };
    return names[phase] || phase || '未知';
}

function App() {
    const [phase, setPhase] = useState<GamePhase>('');
    const [isLCUConnected, setIsLCUConnected] = useState(false);

    useEffect(() => {
        // 初始获取游戏阶段
        GetCurrentPhase().then(p => {
            setPhase(p as GamePhase);
            setIsLCUConnected(true);
        }).catch(() => {
            setIsLCUConnected(false);
        });

        // 监听游戏阶段变化事件
        const unsubscribe = EventsOn("game:phase", (p: string) => {
            setPhase(p as GamePhase);
            setIsLCUConnected(true);
        });

        return () => {
            unsubscribe();
        };
    }, []);

    const renderContent = () => {
        if (!isLCUConnected) {
            return (
                <div className="waiting-screen">
                    <div className="waiting-icon">🎮</div>
                    <h2>等待连接 LOL 客户端</h2>
                    <p>请启动英雄联盟客户端</p>
                </div>
            );
        }

        switch (phase) {
            case 'ChampSelect':
            case 'GameStart':
                return <ChampSelectView />;
            case 'InProgress':
                return (
                    <div className="waiting-screen">
                        <div className="waiting-icon">⚔️</div>
                        <h2>游戏进行中</h2>
                        <p>祝您游戏愉快！</p>
                    </div>
                );
            default:
                return (
                    <div className="waiting-screen">
                        <div className="waiting-icon">🌀</div>
                        <h2>{phaseDisplayName(phase)}</h2>
                        <p>等待进入选人阶段...</p>
                    </div>
                );
        }
    };

    return (
        <div id="App">
            <header className="app-header">
                <div className="app-title">HaxPlugins</div>
                <div className={`phase-badge phase-${phase || 'none'}`}>
                    {phaseDisplayName(phase)}
                </div>
                <div className={`connection-dot ${isLCUConnected ? 'connected' : 'disconnected'}`} />
            </header>
            <main className="app-main">
                {renderContent()}
            </main>
        </div>
    );
}

export default App;
