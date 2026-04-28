import {useEffect, useState} from 'react';
import './App.css';
import {EventsOn} from "../wailsjs/runtime";
import {GetCurrentPhase} from "../wailsjs/go/main/App";
import ChampSelectPanel from "./panels/ChampSelectPanel";

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
        'ChampSelect': '选人中',
        'GameStart': '游戏开始',
        'InProgress': '游戏中',
        'WaitingForStats': '等待结算',
        'PreEndOfGame': '游戏结束',
        'EndOfGame': '已结束',
    };
    return names[phase] || phase || '未知';
}

function phaseDotClass(phase: GamePhase): string {
    if (phase === 'ChampSelect' || phase === 'GameStart') return "bg-lol-green shadow-glow-green";
    if (phase === 'InProgress') return "bg-lol-red shadow-glow-red";
    if (phase === 'ReadyCheck') return "bg-amber-400 shadow-[0_0_8px_rgba(251,191,36,0.4)]";
    return "bg-lol-muted shadow-none";
}

function App() {
    const [phase, setPhase] = useState<GamePhase>('');
    const [isLCUConnected, setIsLCUConnected] = useState(false);

    useEffect(() => {
        GetCurrentPhase().then(p => {
            setPhase(p as GamePhase);
            setIsLCUConnected(true);
        }).catch(() => {
            setIsLCUConnected(false);
        });

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
                <div className="flex flex-col items-center justify-center h-full gap-3">
                    <div className="w-12 h-12 rounded-full bg-lol-card border border-lol-border flex items-center justify-center">
                        <svg className="w-6 h-6 text-lol-muted animate-spin" fill="none" viewBox="0 0 24 24">
                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3"/>
                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
                        </svg>
                    </div>
                    <div className="text-sm text-lol-text font-semibold">等待连接 LOL 客户端</div>
                    <div className="text-xs text-lol-muted">请启动英雄联盟客户端</div>
                </div>
            );
        }

        switch (phase) {
            case 'ChampSelect':
            case 'GameStart':
                return <ChampSelectPanel />;
            case 'InProgress':
                return (
                    <div className="flex flex-col items-center justify-center h-full gap-3">
                        <div className="w-12 h-12 rounded-full bg-lol-card border border-lol-red/30 flex items-center justify-center">
                            <svg className="w-6 h-6 text-lol-red" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"/>
                            </svg>
                        </div>
                        <div className="text-sm text-lol-text font-semibold">游戏进行中</div>
                    </div>
                );
            default:
                return (
                    <div className="flex flex-col items-center justify-center h-full gap-3">
                        <div className="w-12 h-12 rounded-full bg-lol-card border border-lol-border flex items-center justify-center">
                            <svg className="w-6 h-6 text-lol-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"/>
                            </svg>
                        </div>
                        <div className="text-sm text-lol-text">{phaseDisplayName(phase)}</div>
                        <div className="text-xs text-lol-muted">等待进入选人阶段...</div>
                    </div>
                );
        }
    };

    return (
        <div className="flex flex-col h-screen bg-lol-bg-deep text-lol-text font-['Nunito',sans-serif] overflow-hidden select-none">
            {/* Header toolbar */}
            <div className="flex items-center gap-2.5 px-3 py-2 bg-header-gradient border-b border-lol-border/60 flex-shrink-0 drag-region">
                <div className={`w-2 h-2 rounded-full ${phaseDotClass(phase)} transition-all duration-300`} />
                <span className="text-xs font-bold text-gold-shimmer tracking-wider">海克斯大乱斗</span>
                <span className="ml-auto text-[10px] text-lol-muted font-medium">{phaseDisplayName(phase)}</span>
            </div>
            <main className="flex-1 overflow-y-auto">
                {renderContent()}
            </main>
        </div>
    );
}

export default App;