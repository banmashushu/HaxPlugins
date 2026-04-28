import {useEffect, useState, useCallback} from 'react';
import {EventsOn} from "../../wailsjs/runtime";
import {GetMyTeamStats, GetEnemyTeamStats} from "../../wailsjs/go/main/App";
import TeamMemberCard from "./TeamMemberCard";

interface AugmentStat {
    champion_id: number;
    champion_name: string;
    augment_id: string;
    augment_name: string;
    augment_name_cn: string;
    winrate: number;
    pickrate: number;
    tier: string;
    patch: string;
}

interface BuildItem {
    item_id: number;
    name_cn: string;
    slot: number;
    winrate: number;
}

interface Build {
    champion_id: number;
    champion_name: string;
    game_mode: string;
    role: string;
    items: BuildItem[];
    boots: BuildItem | null;
    skill_order: string[];
    runes: string[];
    patch: string;
}

interface SynergyStat {
    champion_id: number;
    champion_name: string;
    synergy_champion_id: number;
    synergy_name: string;
    score_rank: number;
    score: number;
    play: number;
    win: number;
    win_rate: number;
    tier: number;
    game_mode: string;
    patch: string;
}

interface TeamMember {
    champion_id: number;
    champion_name: string;
    champion_name_en: string;
    cell_id: number;
    winrate: number;
    pickrate: number;
    tier: string;
    augments: AugmentStat[];
    build: Build | null;
    synergies: SynergyStat[];
}

function TeamColumn({title, members, accent, subtitle}: {
    title: string;
    members: TeamMember[];
    accent: "blue" | "red";
    subtitle?: string;
}) {
    const isBlue = accent === "blue";
    const borderColor = isBlue ? "border-lol-blue/30" : "border-lol-red/30";
    const bgColor = isBlue ? "bg-lol-blue/5" : "bg-lol-red/5";
    const headerBg = isBlue ? "bg-lol-blue/10" : "bg-lol-red/10";
    const headerBorder = isBlue ? "border-lol-blue/20" : "border-lol-red/20";
    const textColor = isBlue ? "text-lol-blue-bright" : "text-lol-red";
    const badgeBg = isBlue ? "bg-lol-blue-bright" : "bg-lol-red";
    const badgeRing = isBlue ? "ring-lol-blue/40" : "ring-lol-red/40";

    return (
        <div className="flex-1 min-w-0">
            {/* Column Header */}
            <div className={`flex items-center justify-between px-3 py-2 rounded-t-lg border ${headerBorder} ${headerBg}`}>
                <div className="flex items-center gap-2">
                    <span className={`w-2 h-2 rounded-full ${badgeBg} ring-1 ${badgeRing}`}/>
                    <span className={`text-xs font-bold uppercase tracking-widest ${textColor}`}>{title}</span>
                </div>
                {subtitle && (
                    <span className="text-[9px] text-lol-muted font-mono">{subtitle}</span>
                )}
            </div>

            {/* Column Cards */}
            <div className={`border-l border-r border-b ${borderColor} ${bgColor} rounded-b-lg p-2 space-y-1.5`}>
                {members.map((member, idx) => (
                    <TeamMemberCard key={member.cell_id} member={member} accent={accent} position={idx + 1}/>
                ))}
                {members.length === 0 && (
                    <div className="py-8 text-center">
                        <p className="text-[10px] text-lol-muted/50">等待选人中...</p>
                    </div>
                )}
            </div>
        </div>
    );
}

function ChampSelectView() {
    const [team, setTeam] = useState<TeamMember[]>([]);
    const [enemyTeam, setEnemyTeam] = useState<TeamMember[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    const loadData = useCallback(() => {
        setLoading(true);
        setError('');
        Promise.all([
            GetMyTeamStats().catch(err => {
                setError(String(err));
                return [];
            }),
            GetEnemyTeamStats().catch(err => {
                setError(String(err));
                return [];
            }),
        ]).then(([teamStats, enemyStats]) => {
            setTeam(teamStats as TeamMember[]);
            setEnemyTeam(enemyStats as TeamMember[]);
            setLoading(false);
        }).catch(err => {
            setError(String(err));
            setLoading(false);
        });
    }, []);

    useEffect(() => {
        loadData();

        const unsubscribe = EventsOn("game:champselect", () => {
            loadData();
        });

        return () => {
            unsubscribe();
        };
    }, [loadData]);

    if (loading && team.length === 0 && enemyTeam.length === 0) {
        return (
            <div className="flex items-center justify-center py-16 text-lol-muted text-sm">
                <div className="flex items-center gap-2">
                    <svg className="w-4 h-4 animate-spin text-lol-blue" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3"/>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
                    </svg>
                    <span>加载中...</span>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="text-center py-10">
                <div className="w-10 h-10 rounded-full bg-lol-card border border-lol-red/30 flex items-center justify-center mx-auto mb-3">
                    <svg className="w-5 h-5 text-lol-red" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
                    </svg>
                </div>
                <p className="text-lol-red text-sm mb-3">加载失败: {error}</p>
                <button
                    className="px-4 py-1.5 rounded-md bg-lol-card border border-lol-border text-lol-text text-sm hover:border-lol-gold/40 hover:text-lol-gold transition-all duration-200"
                    onClick={loadData}
                >
                    重试
                </button>
            </div>
        );
    }

    if (team.length === 0 && enemyTeam.length === 0) {
        return (
            <div className="text-center py-10">
                <div className="w-10 h-10 rounded-full bg-lol-card border border-lol-border flex items-center justify-center mx-auto mb-3">
                    <svg className="w-5 h-5 text-lol-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72M18 18.72V21m0-2.25a9.094 9.094 0 01-3.741-.479 3 3 0 00-4.682 2.72M18 18.72V18m0 3v-3m0-3v-3"/>
                    </svg>
                </div>
                <p className="text-lol-text text-sm mb-3">暂无队友数据</p>
                <button
                    className="px-4 py-1.5 rounded-md bg-lol-card border border-lol-border text-lol-text text-sm hover:border-lol-gold/40 hover:text-lol-gold transition-all duration-200"
                    onClick={loadData}
                >
                    刷新
                </button>
            </div>
        );
    }

    return (
        <div className="px-3 py-3 max-w-3xl mx-auto">
            {/* Header */}
            <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-1.5">
                    <span className="text-[9px] text-lol-muted uppercase tracking-widest">海克斯大乱斗</span>
                    <span className="text-[9px] text-lol-gold/50 font-bold">5v5</span>
                </div>
                <button
                    className="px-3 py-1 rounded text-[11px] font-semibold text-lol-muted border border-lol-border/50
                               hover:border-lol-gold/30 hover:text-lol-gold transition-all duration-200"
                    onClick={loadData}
                >
                    刷新
                </button>
            </div>

            {/* VS Layout - Two Columns */}
            <div className="flex gap-2">
                <TeamColumn
                    title="我方队伍"
                    members={team}
                    accent="blue"
                    subtitle={`共${team.length}人`}
                />
                <TeamColumn
                    title="敌方队伍"
                    members={enemyTeam}
                    accent="red"
                    subtitle={`共${enemyTeam.length}人`}
                />
            </div>

            <p className="text-[9px] text-lol-muted/60 mt-3 text-center tracking-wide">点击卡片查看详细数据</p>
        </div>
    );
}

export default ChampSelectView;
