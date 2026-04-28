import {useState} from 'react';
import AugmentList from "./AugmentList";
import BuildPanel from "./BuildPanel";
import {championIconURL} from "../utils/ddragon";

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

function winrateColor(winrate: number): string {
    if (winrate >= 55) return "text-lol-green";
    if (winrate >= 50) return "text-lol-gold-bright";
    return "text-lol-red";
}

function winrateBarGradient(winrate: number): string {
    if (winrate >= 55) return "from-lol-green to-emerald-500";
    if (winrate >= 50) return "from-lol-gold-bright to-lol-gold";
    return "from-lol-red to-red-400";
}

function tierBadgeClasses(tier: string): string {
    const t = tier.toLowerCase();
    if (t === "s") return "bg-lol-green/15 text-tier-s ring-1 ring-lol-green/30";
    if (t === "a") return "bg-lol-blue/15 text-tier-a ring-1 ring-lol-blue/30";
    if (t === "b") return "bg-lol-gold/15 text-tier-b ring-1 ring-lol-gold/30";
    if (t === "c") return "bg-lol-gold-dim/20 text-tier-c ring-1 ring-lol-gold-dim/30";
    if (t === "d") return "bg-lol-red/15 text-tier-d ring-1 ring-lol-red/30";
    return "bg-lol-muted/15 text-lol-muted ring-1 ring-lol-muted/30";
}

function synergyWinrateColor(winrate: number): string {
    if (winrate >= 55) return "text-lol-green";
    if (winrate >= 50) return "text-lol-gold-bright";
    return "text-lol-red";
}

function SynergyList({synergies}: { synergies: SynergyStat[] }) {
    if (!synergies || synergies.length === 0) {
        return <p className="text-[10px] text-lol-muted/60">暂无协同数据</p>;
    }

    return (
        <div className="space-y-1">
            {synergies.map((s, idx) => (
                <div key={idx}
                     className="flex items-center justify-between bg-lol-bg/60 rounded-md px-2 py-1 border border-lol-border/30">
                    <div className="flex items-center gap-1.5 min-w-0">
                        <span
                            className="text-[10px] font-bold text-lol-muted w-4 text-center">{s.score_rank}</span>
                        <span className="text-[11px] text-lol-text-bright truncate font-medium">{s.synergy_name}</span>
                    </div>
                    <div className="flex items-center gap-2 flex-shrink-0">
                        <span className={`text-[10px] font-bold ${synergyWinrateColor(s.win_rate)}`}>
                            {s.win_rate.toFixed(1)}%
                        </span>
                        <span className="text-[9px] text-lol-muted font-mono">
                            {s.play >= 1000 ? `${(s.play / 1000).toFixed(1)}k` : s.play}场
                        </span>
                    </div>
                </div>
            ))}
        </div>
    );
}

function TeamMemberCard({member, accent, position}: {
    member: TeamMember;
    accent: "blue" | "red";
    position: number;
}) {
    const [expanded, setExpanded] = useState(false);
    const [activeTab, setActiveTab] = useState<'augments' | 'synergies' | 'build'>('augments');

    const isBlue = accent === "blue";
    const hasData = member.winrate > 0;
    const winratePct = member.winrate * 100;
    const iconURL = championIconURL(member.champion_name_en);
    const accentBorder = isBlue ? "border-lol-blue/30" : "border-lol-red/30";
    const accentBg = isBlue ? "bg-lol-blue/3" : "bg-lol-red/3";
    const accentText = isBlue ? "text-lol-blue-bright" : "text-lol-red";
    const positionBg = isBlue ? "bg-lol-blue/20" : "bg-lol-red/20";

    return (
        <div className={`rounded-md border transition-all duration-200 overflow-hidden ${
            expanded
                ? `${accentBorder} bg-lol-card/80`
                : "border-lol-border/30 bg-lol-card/40 hover:border-lol-border/60"
        }`}>
            {/* Card Header - clickable */}
            <div
                className="flex items-center gap-2.5 px-3 py-2.5 cursor-pointer select-none"
                onClick={() => setExpanded(!expanded)}
            >
                {/* Position Number */}
                <span className={`w-5 h-5 rounded flex items-center justify-center text-[10px] font-extrabold ${positionBg} ${accentText} flex-shrink-0`}>
                    {position}
                </span>

                {/* Champion Avatar */}
                <div className="relative w-9 h-9 rounded-md overflow-hidden ring-1 ring-lol-border/50 flex-shrink-0">
                    {iconURL ? (
                        <img
                            src={iconURL}
                            alt={member.champion_name}
                            className="w-full h-full object-cover"
                            onError={(e) => {
                                const img = e.target as HTMLImageElement;
                                img.style.display = "none";
                                const fallback = img.nextElementSibling as HTMLElement;
                                if (fallback) fallback.classList.remove("hidden");
                            }}
                        />
                    ) : null}
                    <div className={`w-full h-full flex items-center justify-center text-xs font-bold text-lol-muted bg-lol-bg ${iconURL ? "hidden" : ""}`}>
                        {member.champion_name ? member.champion_name.charAt(0) : "?"}
                    </div>
                </div>

                {/* Champion Info */}
                <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-1.5">
                        <span className="text-xs font-bold text-lol-text-bright truncate">
                            {member.champion_name || `英雄 #${member.champion_id}`}
                        </span>
                        {member.tier && (
                            <span className={`px-1 py-0.5 rounded text-[8px] font-bold uppercase ${tierBadgeClasses(member.tier)}`}>
                                {member.tier}
                            </span>
                        )}
                    </div>
                    <div className="flex items-center gap-2 mt-0.5">
                        <div className="flex-1 h-1 rounded-full overflow-hidden stat-bar-bg max-w-[60px]">
                            <div
                                className={`h-full rounded-full bg-gradient-to-r ${winrateBarGradient(winratePct)}`}
                                style={{width: `${Math.min(winratePct, 100)}%`}}
                            />
                        </div>
                        {hasData ? (
                            <span className={`text-[10px] font-bold font-mono ${winrateColor(winratePct)}`}>
                                {winratePct.toFixed(1)}%
                            </span>
                        ) : (
                            <span className="text-[9px] text-lol-muted font-mono">--</span>
                        )}
                    </div>
                </div>

                {/* Expand Arrow */}
                <span className={`text-[10px] text-lol-muted transition-transform duration-200 flex-shrink-0 ${expanded ? "rotate-180" : ""}`}>
                    ▼
                </span>
            </div>

            {/* Expanded Content */}
            {expanded && (
                <div className="border-t border-lol-border/20 px-3 py-2.5 space-y-2">
                    {/* Tabs */}
                    <div className="flex gap-1.5">
                        <button
                            className={`px-2.5 py-1 rounded text-[10px] font-semibold transition-all ${
                                activeTab === 'augments'
                                    ? "bg-lol-gold/20 text-lol-gold-bright ring-1 ring-lol-gold/30"
                                    : "bg-lol-bg/60 text-lol-muted hover:text-lol-text"
                            }`}
                            onClick={() => setActiveTab('augments')}
                        >
                            海克斯推荐
                        </button>
                        <button
                            className={`px-2.5 py-1 rounded text-[10px] font-semibold transition-all ${
                                activeTab === 'synergies'
                                    ? "bg-lol-purple/20 text-lol-purple ring-1 ring-lol-purple/30"
                                    : "bg-lol-bg/60 text-lol-muted hover:text-lol-text"
                            }`}
                            onClick={() => setActiveTab('synergies')}
                        >
                            最佳协同
                        </button>
                        <button
                            className={`px-2.5 py-1 rounded text-[10px] font-semibold transition-all ${
                                activeTab === 'build'
                                    ? "bg-lol-blue/20 text-lol-blue-bright ring-1 ring-lol-blue/30"
                                    : "bg-lol-bg/60 text-lol-muted hover:text-lol-text"
                            }`}
                            onClick={() => setActiveTab('build')}
                        >
                            出装推荐
                        </button>
                    </div>

                    {/* Tab Content */}
                    <div>
                        {activeTab === 'augments' && (
                            <AugmentList augments={member.augments}/>
                        )}
                        {activeTab === 'synergies' && (
                            <SynergyList synergies={member.synergies}/>
                        )}
                        {activeTab === 'build' && (
                            <BuildPanel build={member.build}/>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
}

export default TeamMemberCard;
