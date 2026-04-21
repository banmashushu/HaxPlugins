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

function tierClass(tier: string): string {
    const t = tier.toLowerCase();
    if (t === 'prismatic') return 'tier-prismatic';
    if (t === 'gold') return 'tier-gold';
    if (t === 'silver') return 'tier-silver';
    return 'tier-unknown';
}

function AugmentList({augments}: { augments: AugmentStat[] }) {
    if (!augments || augments.length === 0) {
        return <div className="empty-list">暂无海克斯推荐数据</div>;
    }

    return (
        <div className="augment-list">
            <div className="list-header">
                <span className="col-name">海克斯</span>
                <span className="col-tier">等级</span>
                <span className="col-score">强度</span>
                <span className="col-pick">选取率</span>
            </div>
            {augments.map((a, index) => (
                <div key={a.augment_id} className="augment-row">
                    <span className="col-rank">{index + 1}</span>
                    <span className="col-name">
                        {a.augment_name_cn || a.augment_name || a.augment_id}
                    </span>
                    <span className={`col-tier ${tierClass(a.tier)}`}>
                        {a.tier}
                    </span>
                    <span className="col-score">{a.winrate.toFixed(1)}</span>
                    <span className="col-pick">{(a.pickrate * 100).toFixed(1)}%</span>
                </div>
            ))}
        </div>
    );
}

export default AugmentList;
