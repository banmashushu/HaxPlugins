import { TeamMember, TIER_COLORS, TIER_BORDER_COLORS, formatPercent } from "../shared/types";
import { championIconURL } from "../utils/ddragon";

interface HeroCardProps {
  hero: TeamMember;
  selected: boolean;
  onClick: () => void;
}

export default function HeroCard({ hero, selected, onClick }: HeroCardProps) {
  const tierClass = TIER_COLORS[hero.tier] || 'bg-lol-muted text-white';
  const borderClass = selected
    ? TIER_BORDER_COLORS[hero.tier] || 'border-lol-border-glow'
    : 'border-transparent';

  return (
    <button
      onClick={onClick}
      className={`
        relative flex flex-col items-center p-1.5 rounded-lg
        bg-lol-card/80 border-2 ${borderClass}
        transition-all duration-200 hover:bg-lol-card-hover
        focus:outline-none
        ${selected ? 'scale-105' : 'hover:scale-[1.02]'}
      `}
    >
      {/* Hero avatar */}
      <div className="relative w-14 h-14 rounded-md overflow-hidden">
        <img
          src={championIconURL(hero.champion_name_en)}
          alt={hero.champion_name}
          className="w-full h-full object-cover"
          loading="lazy"
        />
        {/* Tier badge bottom-right */}
        <span
          className={`
            absolute -bottom-0.5 -right-0.5
            text-[10px] font-bold px-1 py-px rounded
            leading-none ${tierClass}
          `}
        >
          {hero.tier || '?'}
        </span>
      </div>

      {/* Name + winrate */}
      <span className="text-[11px] text-lol-text font-semibold mt-1 leading-tight truncate max-w-full">
        {hero.champion_name}
      </span>
      <span className="text-[10px] text-lol-green font-medium">
        {formatPercent(hero.winrate)}
      </span>
    </button>
  );
}
