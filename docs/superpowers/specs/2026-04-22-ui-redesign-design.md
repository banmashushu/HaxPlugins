# UI Redesign & Build Data Design

## Overview

Redesign the HaxPlugins UI to a modern minimal dark style using Tailwind CSS, add champion avatars and item icons from DDragon CDN, and transform the card layout from tab-based to a combined augment+build vertical layout.

## Current State

- 4 React components: ChampSelectView, TeamMemberCard, AugmentList, BuildPanel
- Hand-written CSS in App.css (~470 lines)
- No champion avatars, no item icons — all text-only
- TeamMemberCard uses tabs to switch between Augments and Build
- `build_recommendations` table is empty (0 rows) — scraper not implemented
- `champion_stats` table is empty (0 rows)
- `hero_augment_stats` has 16K+ rows, `augments` has 195, `champions` has 172, `items` has 695

## Design Decisions

### Style: Modern Minimal Dark (Tailwind CSS)

Color palette:
- Background: slate-900 ~ slate-950
- Cards: slate-800 with slate-700 borders
- Accent: amber-400 (gold, LOL-themed)
- Positive: emerald-400 (high winrate)
- Neutral: yellow-400 (medium winrate)
- Negative: red-400 (low winrate)
- Tier colors: purple-400 (prismatic), amber-400 (gold), slate-300 (silver)
- Text primary: slate-100
- Text secondary: slate-400

### Layout Changes

**Header**: Narrower, more transparent feel
**ChampSelectView**: Vertical card list (unchanged structure, improved styling)
**TeamMemberCard**: Remove tab toggle; show augments above, build below in same card
**AugmentList**: Add winrate progress bars, tier color badges
**BuildPanel**: Equipment card grid with icon placeholders, skill order badges

### Champion Avatars

Source: DDragon CDN
URL pattern: `https://ddragon.leagueoflegends.com/cdn/{patch}/img/champion/{key}.png`

New backend method: `GetChampionImageURL(championID int) string`
- Queries `champions` table for `name_en` (used as DDragon key)
- Returns full CDN URL or empty string if not found
- Patch version read from config, fallback to `currentPatch` constant

### Item Icons

Source: DDragon CDN
URL pattern: `https://ddragon.leagueoflegends.com/cdn/{patch}/img/item/{item_id}.png`

Frontend-only: construct URL from item_id, no backend method needed.
Fallback: show item name initial in a colored square.

### Empty Build Data Handling

When `build_recommendations` is empty for a champion:
- Show placeholder text: "暂无出装数据"
- Greyed-out icon placeholders in a 3x2 grid
- Never hide the Build section entirely

## Frontend Changes

### New Files
- `frontend/src/utils/ddragon.ts` — DDragon URL helpers

### Modified Files
- `frontend/package.json` — add tailwindcss, postcss, autoprefixer
- `frontend/postcss.config.js` — PostCSS config
- `frontend/tailwind.config.js` — Tailwind config with custom colors
- `frontend/src/style.css` — replace with Tailwind directives + custom vars
- `frontend/src/App.css` — remove all hand-written styles
- `frontend/src/App.tsx` — minor layout adjustments
- `frontend/src/components/ChampSelectView.tsx` — styling refresh
- `frontend/src/components/TeamMemberCard.tsx` — add avatar, remove tabs, vertical layout
- `frontend/src/components/AugmentList.tsx` — progress bars, tier badges
- `frontend/src/components/BuildPanel.tsx` — item card grid with icons

## Backend Changes

### Modified Files
- `app.go` — add `GetChampionImageURL(championID int) string` method

### No Changes
- Database schema — no migrations needed
- Scraper code — out of scope for this task
- LCU client — no changes