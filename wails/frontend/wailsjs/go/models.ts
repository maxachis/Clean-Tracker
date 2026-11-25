export namespace model {
	
	export class AddChoreRequest {
	    name: string;
	    count: number;

	    static createFrom(source: any = {}) {
	        return new AddChoreRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.count = source["count"];
	    }
	}
	export class Chore {
	    id: string;
	    session_id: string;
	    name: string;
	    target_count: number;
	    position: number;

	    static createFrom(source: any = {}) {
	        return new Chore(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.session_id = source["session_id"];
	        this.name = source["name"];
	        this.target_count = source["target_count"];
	        this.position = source["position"];
	    }
	}
	export class ChoreProgress {
	    id: string;
	    name: string;
	    target: number;
	    done: number;
	    outstanding: number;

	    static createFrom(source: any = {}) {
	        return new ChoreProgress(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.target = source["target"];
	        this.done = source["done"];
	        this.outstanding = source["outstanding"];
	    }
	}
	export class Session {
	    id: string;
	    length_seconds: number;
	    created_at: any;
	    started_at?: any;
	    ended_at?: any;
	    end_status?: string;

	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.length_seconds = source["length_seconds"];
	        this.created_at = source["created_at"];
	        this.started_at = source["started_at"];
	        this.ended_at = source["ended_at"];
	        this.end_status = source["end_status"];
	    }
	}
	export class Draft {
	    session: Session;
	    chores: Chore[];

	    static createFrom(source: any = {}) {
	        return new Draft(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.session = this.convertValues(source["session"], Session);
	        this.chores = this.convertValues(source["chores"], Chore);
	    }

	    convertValues(a: any, classs: any, asMap: boolean = false): any {
	        if (!a) { return a; }
	        if (a.slice && a.map) { return (a as any[]).map(elem => this.convertValues(elem, classs)); }
	        else if ("object" === typeof a) { if (asMap) { for (const key of Object.keys(a)) { a[key] = new classs(a[key]); } return a; } return new classs(a); }
	        return a;
	    }
	}
	export class Snapshot {
	    status: string;
	    chores: ChoreProgress[];
	    done: number;
	    outstanding: number;
	    missed: number;
	    session_seconds_left: number;
	    action_seconds_left: number;
	    action_seconds_total: number;
	    paused: boolean;
	    ended_at?: any;
	    as_of: any;

	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.chores = this.convertValues(source["chores"], ChoreProgress);
	        this.done = source["done"];
	        this.outstanding = source["outstanding"];
	        this.missed = source["missed"];
	        this.session_seconds_left = source["session_seconds_left"];
	        this.action_seconds_left = source["action_seconds_left"];
	        this.action_seconds_total = source["action_seconds_total"];
	        this.paused = source["paused"];
	        this.ended_at = source["ended_at"];
	        this.as_of = source["as_of"];
	    }

	    convertValues(a: any, classs: any, asMap: boolean = false): any {
	        if (!a) { return a; }
	        if (a.slice && a.map) { return (a as any[]).map(elem => this.convertValues(elem, classs)); }
	        else if ("object" === typeof a) { if (asMap) { for (const key of Object.keys(a)) { a[key] = new classs(a[key]); } return a; } return new classs(a); }
	        return a;
	    }
	}
	export class SessionView {
	    session_id: string;
	    session: Session;
	    snapshot: Snapshot;

	    static createFrom(source: any = {}) {
	        return new SessionView(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.session_id = source["session_id"];
	        this.session = this.convertValues(source["session"], Session);
	        this.snapshot = this.convertValues(source["snapshot"], Snapshot);
	    }

	    convertValues(a: any, classs: any, asMap: boolean = false): any {
	        if (!a) { return a; }
	        if (a.slice && a.map) { return (a as any[]).map(elem => this.convertValues(elem, classs)); }
	        else if ("object" === typeof a) { if (asMap) { for (const key of Object.keys(a)) { a[key] = new classs(a[key]); } return a; } return new classs(a); }
	        return a;
	    }
	}
	export class Settings {
	    last_session_minutes: number;
	    miss_sound: boolean;

	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.last_session_minutes = source["last_session_minutes"];
	        this.miss_sound = source["miss_sound"];
	    }
	}
	export class Summary {
	    session: Session;
	    snapshot: Snapshot;

	    static createFrom(source: any = {}) {
	        return new Summary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.session = this.convertValues(source["session"], Session);
	        this.snapshot = this.convertValues(source["snapshot"], Snapshot);
	    }

	    convertValues(a: any, classs: any, asMap: boolean = false): any {
	        if (!a) { return a; }
	        if (a.slice && a.map) { return (a as any[]).map(elem => this.convertValues(elem, classs)); }
	        else if ("object" === typeof a) { if (asMap) { for (const key of Object.keys(a)) { a[key] = new classs(a[key]); } return a; } return new classs(a); }
	        return a;
	    }
	}
	export class UpdateChoreRequest {
	    id: string;
	    name: string;
	    count: number;

	    static createFrom(source: any = {}) {
	        return new UpdateChoreRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.count = source["count"];
	    }
	}
	export class UpdateSettingsRequest {
	    last_session_minutes: number;
	    miss_sound: boolean;

	    static createFrom(source: any = {}) {
	        return new UpdateSettingsRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.last_session_minutes = source["last_session_minutes"];
	        this.miss_sound = source["miss_sound"];
	    }
	}

}

