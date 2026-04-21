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

function BuildPanel({build}: { build: Build | null }) {
    if (!build) {
        return <div className="empty-list">暂无出装推荐数据</div>;
    }

    return (
        <div className="build-panel">
            {build.items && build.items.length > 0 && (
                <div className="build-section">
                    <h4>核心装备</h4>
                    <div className="items-grid">
                        {build.items.map((item, idx) => (
                            <div key={idx} className="item-slot">
                                <div className="item-icon">{item.name_cn.charAt(0)}</div>
                                <div className="item-name">{item.name_cn}</div>
                                <div className="item-winrate">{item.winrate.toFixed(1)}%</div>
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {build.boots && (
                <div className="build-section">
                    <h4>鞋子</h4>
                    <div className="item-slot boots">
                        <div className="item-icon">👢</div>
                        <div className="item-name">{build.boots.name_cn}</div>
                    </div>
                </div>
            )}

            {build.skill_order && build.skill_order.length > 0 && (
                <div className="build-section">
                    <h4>技能加点</h4>
                    <div className="skill-order">
                        {build.skill_order.map((skill, idx) => (
                            <span key={idx} className="skill-badge">{skill}</span>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
}

export default BuildPanel;
