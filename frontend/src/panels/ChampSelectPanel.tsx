import { useEffect, useState, useCallback } from "react";
import { EventsOn } from "../../wailsjs/runtime";
import { GetMyTeamStats } from "../../wailsjs/go/main/App";
import { TeamMember } from "../shared/types";
import HeroGrid from "./HeroGrid";
import HeroDetail from "./HeroDetail";

export default function ChampSelectPanel() {
  const [heroes, setHeroes] = useState<TeamMember[]>([]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await GetMyTeamStats();
      setHeroes((data || []).map((d: TeamMember | string) => typeof d === 'string' ? JSON.parse(d) : d));
      setSelectedId((prev) => {
        if (prev != null || !data || data.length === 0) return prev;
        return data[0].champion_id;
      });
    } catch {
      setError('无法获取队伍数据');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
    const unsub = EventsOn("game:champselect", () => loadData());
    return () => unsub();
  }, [loadData]);

  const selectedHero = heroes.find((h) => h.champion_id === selectedId) || null;

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="flex flex-col items-center gap-2">
          <div className="w-6 h-6 border-2 border-lol-gold/30 border-t-lol-gold rounded-full animate-spin" />
          <span className="text-xs text-lol-muted">加载队伍数据...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-xs text-lol-red">{error}</div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex-shrink-0 py-3">
        <HeroGrid
          heroes={heroes}
          selectedId={selectedId}
          onSelect={(h) => setSelectedId(h.champion_id)}
        />
      </div>

      <div className="mx-3 border-t border-lol-border/30" />

      <div className="flex-1 overflow-y-auto py-2">
        <HeroDetail hero={selectedHero} />
      </div>
    </div>
  );
}
