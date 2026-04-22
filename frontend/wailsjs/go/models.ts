export namespace data {
	
	export class BuildItem {
	    item_id: number;
	    name_cn: string;
	    slot: number;
	    winrate: number;
	
	    static createFrom(source: any = {}) {
	        return new BuildItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.item_id = source["item_id"];
	        this.name_cn = source["name_cn"];
	        this.slot = source["slot"];
	        this.winrate = source["winrate"];
	    }
	}
	export class Build {
	    champion_id: number;
	    champion_name: string;
	    game_mode: string;
	    role: string;
	    items: BuildItem[];
	    boots?: BuildItem;
	    skill_order: string[];
	    runes: string[];
	    patch: string;
	
	    static createFrom(source: any = {}) {
	        return new Build(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.champion_id = source["champion_id"];
	        this.champion_name = source["champion_name"];
	        this.game_mode = source["game_mode"];
	        this.role = source["role"];
	        this.items = this.convertValues(source["items"], BuildItem);
	        this.boots = this.convertValues(source["boots"], BuildItem);
	        this.skill_order = source["skill_order"];
	        this.runes = source["runes"];
	        this.patch = source["patch"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ChampionSynergy {
	    champion_id: number;
	    champion_name: string;
	    synergy_champion_id: number;
	    synergy_name: string;
	    score_rank: number;
	    score: number;
	    play: number;
	    win: number;
	    win_rate: number;
	    tier: number;
	    game_mode: string;
	    patch: string;
	
	    static createFrom(source: any = {}) {
	        return new ChampionSynergy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.champion_id = source["champion_id"];
	        this.champion_name = source["champion_name"];
	        this.synergy_champion_id = source["synergy_champion_id"];
	        this.synergy_name = source["synergy_name"];
	        this.score_rank = source["score_rank"];
	        this.score = source["score"];
	        this.play = source["play"];
	        this.win = source["win"];
	        this.win_rate = source["win_rate"];
	        this.tier = source["tier"];
	        this.game_mode = source["game_mode"];
	        this.patch = source["patch"];
	    }
	}
	export class HeroAugmentStat {
	    champion_id: number;
	    champion_name: string;
	    augment_id: string;
	    augment_name: string;
	    augment_name_cn: string;
	    winrate: number;
	    pickrate: number;
	    tier: string;
	    patch: string;
	
	    static createFrom(source: any = {}) {
	        return new HeroAugmentStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.champion_id = source["champion_id"];
	        this.champion_name = source["champion_name"];
	        this.augment_id = source["augment_id"];
	        this.augment_name = source["augment_name"];
	        this.augment_name_cn = source["augment_name_cn"];
	        this.winrate = source["winrate"];
	        this.pickrate = source["pickrate"];
	        this.tier = source["tier"];
	        this.patch = source["patch"];
	    }
	}

}

export namespace main {
	
	export class ChampionImage {
	    image_url: string;
	    name_en: string;
	    champion_id: number;
	
	    static createFrom(source: any = {}) {
	        return new ChampionImage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.image_url = source["image_url"];
	        this.name_en = source["name_en"];
	        this.champion_id = source["champion_id"];
	    }
	}
	export class TeamMemberStats {
	    champion_id: number;
	    champion_name: string;
	    champion_name_en: string;
	    cell_id: number;
	    winrate: number;
	    pickrate: number;
	    tier: string;
	    augments: data.HeroAugmentStat[];
	    build?: data.Build;
	    synergies: data.ChampionSynergy[];
	
	    static createFrom(source: any = {}) {
	        return new TeamMemberStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.champion_id = source["champion_id"];
	        this.champion_name = source["champion_name"];
	        this.champion_name_en = source["champion_name_en"];
	        this.cell_id = source["cell_id"];
	        this.winrate = source["winrate"];
	        this.pickrate = source["pickrate"];
	        this.tier = source["tier"];
	        this.augments = this.convertValues(source["augments"], data.HeroAugmentStat);
	        this.build = this.convertValues(source["build"], data.Build);
	        this.synergies = this.convertValues(source["synergies"], data.ChampionSynergy);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

