import { TeamMember } from "../shared/types";
import HeroCard from "./HeroCard";

interface HeroGridProps {
  heroes: TeamMember[];
  selectedId: number | null;
  onSelect: (hero: TeamMember) => void;
}

export default function HeroGrid({ heroes, selectedId, onSelect }: HeroGridProps) {
  if (heroes.length === 0) {
    return (
      <div className="flex items-center justify-center py-8 text-xs text-lol-muted">
        等待英雄数据...
      </div>
    );
  }

  return (
    <div className="grid grid-cols-3 gap-2 px-3">
      {heroes.map((hero) => (
        <HeroCard
          key={hero.cell_id || hero.champion_id}
          hero={hero}
          selected={hero.champion_id === selectedId}
          onClick={() => onSelect(hero)}
        />
      ))}
    </div>
  );
}
