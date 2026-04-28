import { useState } from "react";
import {
  TeamMember, HeroAugmentStat, Build, ChampionSynergy,
  TIER_COLORS, formatPercent,
} from "../shared/types";
import { championIconURL, itemIconURL } from "../utils/ddragon";

type TabId = 'augments' | 'build' | 'synergies';

interface HeroDetailProps {
  hero: TeamMember | null;
}

export default function HeroDetail({ hero }: HeroDetailProps) {
  const [tab, setTab] = useState<TabId>('augments');

  if (!hero) {
    return (
      <div className="flex items-center justify-center py-6 text-xs text-lol-muted">
        点击英雄查看详情
      </div>
    );
  }

  const tierClass = TIER_COLORS[hero.tier] || '';

  return (
    <div className="px-3 pb-3 animate-[slideUp_0.25s_ease-out]">
      {/* Hero header */}
      <div className="flex items-center gap-3 mb-3">
        <img
          src={championIconURL(hero.champion_name_en)}
          alt={hero.champion_name}
          className="w-12 h-12 rounded-lg border border-lol-border/50"
        />
        <div>
          <div className="flex items-center gap-2">
            <span className="text-sm font-bold text-lol-text-bright">
              {hero.champion_name}
            </span>
            <span className={`text-[10px] font-bold px-1.5 py-0.5 rounded leading-none ${tierClass}`}>
              {hero.tier || '?'}级
            </span>
          </div>
          <div className="flex gap-3 mt-0.5">
            <span className="text-xs text-lol-green font-semibold">
              胜率 {formatPercent(hero.winrate)}
            </span>
            <span className="text-xs text-lol-muted">
              选取率 {formatPercent(hero.pickrate)}
            </span>
          </div>
        </div>
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 mb-2 border-b border-lol-border/30">
        {([
          ['augments', '海克斯'],
          ['build', '出装'],
          ['synergies', '协同'],
        ] as [TabId, string][]).map(([id, label]) => (
          <button
            key={id}
            onClick={() => setTab(id)}
            className={`
              text-xs font-semibold px-2.5 py-1.5 rounded-t
              transition-colors duration-150
              ${tab === id
                ? 'bg-lol-card text-lol-gold border-b-2 border-lol-gold'
                : 'text-lol-muted hover:text-lol-text'}
            `}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div className="min-h-[60px]">
        {tab === 'augments' && <AugmentTab augments={hero.augments || []} />}
        {tab === 'build' && <BuildTab build={hero.build} />}
        {tab === 'synergies' && <SynergyTab synergies={hero.synergies || []} />}
      </div>
    </div>
  );
}

function AugmentTab({ augments }: { augments: HeroAugmentStat[] }) {
  if (augments.length === 0) {
    return <div className="text-xs text-lol-muted py-2">暂无海克斯推荐数据</div>;
  }

  const top3 = augments.slice(0, 3);

  return (
    <div className="space-y-1.5">
      {top3.map((a, i) => {
        const tierClass = TIER_COLORS[a.tier] || 'bg-lol-muted text-white';
        return (
          <div
            key={a.augment_id}
            className="flex items-center justify-between px-2 py-1.5 rounded bg-lol-card/50 border border-lol-border/20"
          >
            <div className="flex items-center gap-2">
              <span className="text-[10px] font-bold text-lol-gold w-4">
                #{i + 1}
              </span>
              <span className="text-xs text-lol-text">{a.augment_name_cn || a.augment_name}</span>
            </div>
            <div className="flex items-center gap-2">
              <span className={`text-[10px] font-bold px-1 rounded leading-none ${tierClass}`}>
                {a.tier || '?'}
              </span>
              <span className="text-xs text-lol-green font-semibold w-12 text-right">
                {formatPercent(a.winrate)}
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function BuildTab({ build }: { build?: Build }) {
  if (!build) {
    return <div className="text-xs text-lol-muted py-2">暂无出装推荐数据</div>;
  }

  const items = [...(build.items || [])];
  if (build.boots) items.push(build.boots);

  return (
    <div className="space-y-2">
      <div className="flex gap-1.5 flex-wrap">
        {items.map((item) => (
          <div
            key={item.item_id}
            className="relative w-8 h-8 rounded bg-lol-card/60 border border-lol-border/30 overflow-hidden"
            title={`${item.name_cn} (${formatPercent(item.winrate)})`}
          >
            <img
              src={itemIconURL(item.item_id)}
              alt={item.name_cn}
              className="w-full h-full object-cover"
              loading="lazy"
            />
          </div>
        ))}
      </div>
      {build.skill_order && build.skill_order.length > 0 && (
        <div className="flex items-center gap-1">
          <span className="text-[10px] text-lol-muted">技能:</span>
          {build.skill_order.map((s: string, i: number) => (
            <span key={i} className="text-[11px] font-bold text-lol-text-bright bg-lol-card/60 px-1 rounded">
              {s}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

function SynergyTab({ synergies }: { synergies: ChampionSynergy[] }) {
  if (synergies.length === 0) {
    return <div className="text-xs text-lol-muted py-2">暂无协同推荐数据</div>;
  }

  return (
    <div className="space-y-1.5">
      {synergies.map((s) => (
        <div
          key={s.synergy_champion_id}
          className="flex items-center justify-between px-2 py-1.5 rounded bg-lol-card/50 border border-lol-border/20"
        >
          <div className="flex items-center gap-2">
            <img
              src={championIconURL(s.synergy_name)}
              alt={s.synergy_name}
              className="w-6 h-6 rounded"
              loading="lazy"
            />
            <span className="text-xs text-lol-text">{s.synergy_name}</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-[10px] text-lol-muted">
              {s.play}场
            </span>
            <span className="text-xs text-lol-green font-semibold w-12 text-right">
              {formatPercent(s.win_rate)}
            </span>
          </div>
        </div>
      ))}
    </div>
  );
}
