import {itemIconURL} from "../utils/ddragon";

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

function ItemCard({item}: { item: BuildItem }) {
    const iconURL = itemIconURL(item.item_id);
    return (
        <div className="flex items-center gap-2 bg-lol-bg/60 rounded-md px-2 py-1.5 border border-lol-border/30 hover:border-lol-border-glow/30 transition-colors">
            <div className="w-7 h-7 rounded overflow-hidden flex-shrink-0 bg-lol-card ring-1 ring-lol-border/40">
                <img
                    src={iconURL}
                    alt={item.name_cn}
                    className="w-full h-full object-cover"
                    onError={(e) => {
                        const img = e.target as HTMLImageElement;
                        img.style.display = "none";
                    }}
                />
            </div>
            <div className="flex-1 min-w-0">
                <div className="text-[11px] text-lol-text-bright truncate font-medium">{item.name_cn}</div>
                {item.winrate > 0 && (
                    <div className="text-[9px] text-lol-muted font-mono">{item.winrate.toFixed(1)}%</div>
                )}
            </div>
        </div>
    );
}

function BuildPanel({build}: { build: Build | null }) {
    if (!build) {
        return (
            <div className="py-2 text-center">
                <div className="grid grid-cols-3 gap-1.5 mb-2">
                    {Array.from({length: 6}).map((_, i) => (
                        <div key={i}
                             className="h-9 rounded-md bg-lol-card/50 border border-dashed border-lol-border/40"/>
                    ))}
                </div>
                <p className="text-[10px] text-lol-muted/60">暂无出装推荐数据</p>
            </div>
        );
    }

    return (
        <div className="space-y-2">
            {/* Core Items */}
            {build.items && build.items.length > 0 && (
                <div className="grid grid-cols-2 gap-1">
                    {build.items.map((item, idx) => (
                        <ItemCard key={idx} item={item}/>
                    ))}
                </div>
            )}

            {/* Boots */}
            {build.boots && (
                <ItemCard item={build.boots}/>
            )}

            {/* Skill Order */}
            {build.skill_order && build.skill_order.length > 0 && (
                <div>
                    <div className="text-[9px] text-lol-muted mb-1 uppercase tracking-wider font-semibold">技能加点</div>
                    <div className="flex gap-0.5 flex-wrap">
                        {build.skill_order.map((skill, idx) => (
                            <span
                                key={idx}
                                className={`inline-flex items-center justify-center w-5 h-5 rounded text-[10px] font-bold ${
                                    skill.toUpperCase() === "Q"
                                        ? "bg-lol-blue/20 text-lol-blue-bright ring-1 ring-lol-blue/30"
                                        : skill.toUpperCase() === "W"
                                            ? "bg-lol-green/20 text-lol-green ring-1 ring-lol-green/30"
                                            : skill.toUpperCase() === "E"
                                                ? "bg-lol-gold/15 text-lol-gold-bright ring-1 ring-lol-gold/30"
                                                : "bg-lol-purple/15 text-lol-purple ring-1 ring-lol-purple/30"
                                }`}
                            >
                                {skill.toUpperCase()}
                            </span>
                        ))}
                    </div>
                </div>
            )}

            {/* Runes */}
            {build.runes && build.runes.length > 0 && (
                <div>
                    <div className="text-[9px] text-lol-muted mb-1 uppercase tracking-wider font-semibold">符文</div>
                    <div className="flex flex-wrap gap-1">
                        {build.runes.map((rune, idx) => (
                            <span
                                key={idx}
                                className={`text-[10px] px-1.5 py-0.5 rounded font-medium ${
                                    idx === 0
                                        ? "bg-lol-gold/20 text-lol-gold-bright ring-1 ring-lol-gold/40"
                                        : idx < 4
                                            ? "bg-lol-bg/80 text-lol-text ring-1 ring-lol-border/40"
                                            : idx < 6
                                                ? "bg-lol-blue/10 text-lol-blue-bright ring-1 ring-lol-blue/30"
                                                : "bg-lol-muted/10 text-lol-muted ring-1 ring-lol-muted/20"
                                }`}
                                title={idx === 0 ? "基石符文" : idx < 4 ? "主系" : idx < 6 ? "副系" : "属性"}
                            >
                                {rune}
                            </span>
                        ))}
                    </div>
                </div>
            )}

            {/* No items at all */}
            {(!build.items || build.items.length === 0) && !build.boots && (
                <p className="text-[10px] text-lol-muted/60">暂无出装数据</p>
            )}
        </div>
    );
}

export default BuildPanel;