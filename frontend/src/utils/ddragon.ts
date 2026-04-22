const DD_BASE = "https://ddragon.leagueoflegends.com/cdn";
const PATCH = "16.8.1";

export function championIconURL(nameEN: string): string {
  if (!nameEN) return "";
  return `${DD_BASE}/${PATCH}/img/champion/${nameEN}.png`;
}

export function itemIconURL(itemID: number): string {
  if (!itemID) return "";
  return `${DD_BASE}/${PATCH}/img/item/${itemID}.png`;
}