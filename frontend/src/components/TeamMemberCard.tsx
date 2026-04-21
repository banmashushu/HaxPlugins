import {useState} from 'react';
import AugmentList from "./AugmentList";
import BuildPanel from "./BuildPanel";

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

interface TeamMember {
    champion_id: number;
    champion_name: string;
    cell_id: number;
    winrate: number;
    pickrate: number;
    tier: string;
    augments: AugmentStat[];
    build: Build | null;
}

function winrateColor(winrate: number): string {
    if (winrate >= 55) return 'high';
    if (winrate >= 50) return 'medium';
    return 'low';
}

function TeamMemberCard({member}: { member: TeamMember }) {
    const [expanded, setExpanded] = useState(false);
    const [activeTab, setActiveTab] = useState<'augments' | 'build'>('augments');

    const winrateClass = winrateColor(member.winrate);
    const hasData = member.winrate > 0;

    return (
        <div className={`team-member-card ${expanded ? 'expanded' : ''}`}>
            <div className="card-header" onClick={() => setExpanded(!expanded)}>
                <div className="champion-info">
                    <div className="champion-name">
                        {member.champion_name || `英雄 #${member.champion_id}`}
                    </div>
                    {member.tier && (
                        <span className={`tier-tag tier-${member.tier.toLowerCase()}`}>
                            {member.tier}
                        </span>
                    )}
                </div>
                <div className="winrate-section">
                    {hasData ? (
                        <>
                            <div className={`winrate-value ${winrateClass}`}>
                                {member.winrate.toFixed(1)}%
                            </div>
                            <div className="winrate-label">胜率</div>
                        </>
                    ) : (
                        <div className="no-data">暂无数据</div>
                    )}
                </div>
                <div className="expand-arrow">{expanded ? '▼' : '▶'}</div>
            </div>

            {expanded && (
                <div className="card-details">
                    <div className="detail-tabs">
                        <button
                            className={`tab-btn ${activeTab === 'augments' ? 'active' : ''}`}
                            onClick={() => setActiveTab('augments')}
                        >
                            海克斯推荐
                        </button>
                        <button
                            className={`tab-btn ${activeTab === 'build' ? 'active' : ''}`}
                            onClick={() => setActiveTab('build')}
                        >
                            出装推荐
                        </button>
                    </div>
                    <div className="detail-content">
                        {activeTab === 'augments' && (
                            <AugmentList augments={member.augments} />
                        )}
                        {activeTab === 'build' && (
                            <BuildPanel build={member.build} />
                        )}
                    </div>
                </div>
            )}
        </div>
    );
}

export default TeamMemberCard;
