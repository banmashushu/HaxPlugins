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

function tierBadgeClasses(tier: string): string {
    const t = tier.toLowerCase();
    if (t === "prismatic") return "bg-lol-purple/20 text-tier-prismatic ring-1 ring-lol-purple/30 tier-s";
    if (t === "gold") return "bg-lol-gold/15 text-tier-gold ring-1 ring-lol-gold/30 tier-a";
    if (t === "silver") return "bg-lol-text/10 text-tier-silver ring-1 ring-lol-text/20";
    return "bg-lol-card text-lol-muted ring-1 ring-lol-border/50";
}

function winrateBarGradient(winrate: number): string {
    if (winrate >= 55) return "from-lol-green to-emerald-400";
    if (winrate >= 50) return "from-lol-gold-bright to-lol-gold";
    return "from-lol-red to-red-400";
}

function AugmentList({augments}: { augments: AugmentStat[] }) {
    if (!augments || augments.length === 0) {
        return <div className="text-xs text-lol-muted py-2">暂无海克斯推荐数据</div>;
    }

    return (
        <div className="space-y-0.5">
            {augments.map((a, index) => (
                <div
                    key={a.augment_id}
                    className="flex items-center gap-2 px-2 py-1.5 rounded-md hover:bg-white/[0.03] transition-colors"
                >
                    {/* Rank */}
                    <span className="w-4 text-center text-[10px] font-mono text-lol-muted/70">
                        {index + 1}
                    </span>

                    {/* Tier badge */}
                    <span
                        className={`px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wide ${tierBadgeClasses(a.tier)}`}
                    >
                        {a.tier}
                    </span>

                    {/* Name */}
                    <span className="flex-1 text-[11px] text-lol-text-bright truncate">
                        {a.augment_name_cn || a.augment_name || a.augment_id}
                    </span>

                    {/* Winrate bar */}
                    <div className="w-14 flex items-center gap-1.5">
                        <div className="flex-1 h-1.5 rounded-full overflow-hidden stat-bar-bg">
                            <div
                                className={`h-full rounded-full bg-gradient-to-r ${winrateBarGradient(a.winrate)}`}
                                style={{width: `${Math.min(a.winrate, 100)}%`}}
                            />
                        </div>
                        <span className={`text-[10px] font-mono font-semibold w-7 text-right ${
                            a.winrate >= 55 ? "text-lol-green" : a.winrate >= 50 ? "text-lol-gold-bright" : "text-lol-red"
                        }`}>
                            {a.winrate.toFixed(1)}
                        </span>
                    </div>

                    {/* Pickrate */}
                    <span className="text-[9px] font-mono text-lol-muted/60 w-9 text-right">
                        {(a.pickrate * 100).toFixed(1)}%
                    </span>
                </div>
            ))}
        </div>
    );
}

export default AugmentList;