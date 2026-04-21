import {useEffect, useState, useCallback} from 'react';
import {EventsOn} from "../../wailsjs/runtime";
import {GetMyTeamStats} from "../../wailsjs/go/main/App";
import TeamMemberCard from "./TeamMemberCard";

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

function ChampSelectView() {
    const [team, setTeam] = useState<TeamMember[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    const loadTeam = useCallback(() => {
        setLoading(true);
        setError('');
        GetMyTeamStats().then(stats => {
            setTeam(stats as TeamMember[]);
            setLoading(false);
        }).catch(err => {
            setError(String(err));
            setLoading(false);
        });
    }, []);

    useEffect(() => {
        loadTeam();

        // 监听选人会话更新事件，自动刷新
        const unsubscribe = EventsOn("game:champselect", () => {
            loadTeam();
        });

        return () => {
            unsubscribe();
        };
    }, [loadTeam]);

    if (loading && team.length === 0) {
        return (
            <div className="champ-select-view">
                <div className="loading-spinner">加载中...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="champ-select-view">
                <div className="error-message">加载失败: {error}</div>
                <button className="btn" onClick={loadTeam}>重试</button>
            </div>
        );
    }

    if (team.length === 0) {
        return (
            <div className="champ-select-view">
                <div className="empty-message">暂无队友数据</div>
                <button className="btn" onClick={loadTeam}>刷新</button>
            </div>
        );
    }

    return (
        <div className="champ-select-view">
            <div className="champ-select-header">
                <h3>我方队伍</h3>
                <button className="btn btn-small" onClick={loadTeam}>刷新</button>
            </div>
            <div className="team-grid">
                {team.map(member => (
                    <TeamMemberCard key={member.cell_id} member={member} />
                ))}
            </div>
        </div>
    );
}

export default ChampSelectView;
