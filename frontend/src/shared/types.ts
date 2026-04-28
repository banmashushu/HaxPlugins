import { data, main } from "../../wailsjs/go/models";

export type TeamMember = main.TeamMemberStats;
export type HeroAugmentStat = data.HeroAugmentStat;
export type Build = data.Build;
export type ChampionSynergy = data.ChampionSynergy;
export type BuildItem = data.BuildItem;

export type TierLabel = 'S' | 'A' | 'B' | 'C' | 'D' | '';

export const TIER_COLORS: Record<string, string> = {
  S: 'bg-tier-s text-black',
  A: 'bg-tier-a text-white',
  B: 'bg-tier-b text-white',
  C: 'bg-tier-c text-black',
  D: 'bg-tier-d text-white',
};

export const TIER_BORDER_COLORS: Record<string, string> = {
  S: 'border-tier-s shadow-[0_0_12px_rgba(200,170,110,0.3)]',
  A: 'border-tier-a shadow-[0_0_12px_rgba(180,142,240,0.3)]',
  B: 'border-tier-b shadow-[0_0_12px_rgba(3,151,171,0.3)]',
  C: 'border-tier-c shadow-[0_0_12px_rgba(45,212,160,0.3)]',
  D: 'border-tier-d shadow-[0_0_12px_rgba(90,106,126,0.3)]',
};

export function formatPercent(n: number): string {
  if (n == null) return '--';
  return (n * 100).toFixed(1) + '%';
}
