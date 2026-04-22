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

function tierClasses(tier: string): string {
    const t = tier.toLowerCase();
    if (t === "s") return "bg-lol-green/15 text-tier-s ring-1 ring-lol-green/30 tier-s";
    if (t === "a") return "bg-lol-blue/15 text-tier-a ring-1 ring-lol-blue/30 tier-a";
    if (t === "b") return "bg-lol-gold/15 text-tier-b ring-1 ring-lol-gold/30 tier-b";
    if (t === "c") return "bg-lol-gold-dim/20 text-tier-c ring-1 ring-lol-gold-dim/30 tier-c";
    if (t === "d") return "bg-lol-red/15 text-tier-d ring-1 ring-lol-red/30 tier-d";
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

function TeamMemberCard({member}: { member: TeamMember }) {
    const hasData = member.winrate > 0;
    const winratePct = member.winrate * 100;
    const iconURL = championIconURL(member.champion_name_en);

    return (
        <div className="group relative">
            {/* Compact column — always visible */}
            <div className="glass-card flex flex-col items-center gap-2 p-3 rounded-lg transition-all duration-200 cursor-default">
                {/* Champion Avatar */}
                <div className="relative w-12 h-12 rounded-lg overflow-hidden ring-1 ring-lol-border/80 group-hover:ring-lol-gold/40 transition-all duration-300">
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
                    <div className={`w-full h-full flex items-center justify-center text-base font-bold text-lol-muted bg-lol-card ${iconURL ? "hidden" : ""}`}>
                        {member.champion_name ? member.champion_name.charAt(0) : "?"}
                    </div>
                    {/* Tier badge overlay on avatar bottom */}
                    {member.tier && (
                        <div className="absolute -bottom-0.5 left-1/2 -translate-x-1/2">
                            <span className={`px-1.5 py-px rounded text-[9px] font-bold uppercase tracking-wider ${tierClasses(member.tier)}`}>
                                {member.tier}
                            </span>
                        </div>
                    )}
                </div>

                {/* Champion Name */}
                <div className="text-[11px] font-bold text-lol-text-bright text-center truncate w-full leading-tight">
                    {member.champion_name || `#${member.champion_id}`}
                </div>

                {/* Winrate */}
                <div className="text-center w-full">
                    {hasData ? (
                        <>
                            <div className={`text-sm font-bold leading-tight ${winrateColor(winratePct)}`}>
                                {winratePct.toFixed(1)}%
                            </div>
                            <div className="w-full h-1.5 rounded-full mt-1 overflow-hidden stat-bar-bg">
                                <div
                                    className={`h-full rounded-full bg-gradient-to-r ${winrateBarGradient(winratePct)}`}
                                    style={{width: `${Math.min(winratePct, 100)}%`}}
                                />
                            </div>
                        </>
                    ) : (
                        <div className="text-[10px] text-lol-muted">暂无数据</div>
                    )}
                </div>
            </div>

            {/* Hover detail popup */}
            <div className="absolute left-1/2 -translate-x-1/2 top-full mt-2 z-50 w-72
                            opacity-0 invisible group-hover:opacity-100 group-hover:visible
                            transition-all duration-200 pointer-events-none group-hover:pointer-events-auto">
                <div className="popup-panel rounded-lg p-3.5 space-y-3">
                    <div className="flex items-center gap-2.5 pb-2.5 border-b border-lol-border/40">
                        <div className="w-9 h-9 rounded-md overflow-hidden ring-1 ring-lol-border/60">
                            {iconURL ? (
                                <img src={iconURL} alt={member.champion_name}
                                     className="w-full h-full object-cover"
                                     onError={(e) => {
                                         const img = e.target as HTMLImageElement;
                                         img.style.display = "none";
                                     }} />
                            ) : null}
                        </div>
                        <div className="flex-1 min-w-0">
                            <div className="text-sm font-bold text-lol-text-bright truncate">
                                {member.champion_name || `英雄 #${member.champion_id}`}
                            </div>
                            {hasData && (
                                <div className={`text-xs font-semibold ${winrateColor(winratePct)}`}>
                                    胜率 {winratePct.toFixed(1)}%
                                </div>
                            )}
                        </div>
                        {member.tier && (
                            <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase ${tierClasses(member.tier)}`}>
                                {member.tier}
                            </span>
                        )}
                    </div>

                    <div>
                        <div className="text-[10px] font-bold text-lol-gold/80 uppercase tracking-wider mb-1.5">海克斯推荐</div>
                        <AugmentList augments={member.augments}/>
                    </div>

                    <div>
                        <div className="text-[10px] font-bold text-lol-purple/80 uppercase tracking-wider mb-1.5">最佳协同</div>
                        <SynergyList synergies={member.synergies}/>
                    </div>

                    <div>
                        <div className="text-[10px] font-bold text-lol-blue-bright/80 uppercase tracking-wider mb-1.5">出装推荐</div>
                        <BuildPanel build={member.build}/>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default TeamMemberCard;