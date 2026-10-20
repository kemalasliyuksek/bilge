export namespace cleanup {
	
	export class Candidate {
	    kind: string;
	    id: string;
	    name: string;
	    verdict: string;
	    reason: string;
	    project: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new Candidate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.id = source["id"];
	        this.name = source["name"];
	        this.verdict = source["verdict"];
	        this.reason = source["reason"];
	        this.project = source["project"];
	        this.size = source["size"];
	    }
	}
	export class Report {
	    candidates: Candidate[];
	
	    static createFrom(source: any = {}) {
	        return new Report(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.candidates = this.convertValues(source["candidates"], Candidate);
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

export namespace inventory {
	
	export class BuildCacheRecord {
	    id: string;
	    type: string;
	    inUse: boolean;
	    shared: boolean;
	    size: number;
	    // Go type: time
	    lastUsedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new BuildCacheRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.inUse = source["inUse"];
	        this.shared = source["shared"];
	        this.size = source["size"];
	        this.lastUsedAt = this.convertValues(source["lastUsedAt"], null);
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
	export class Container {
	    id: string;
	    name: string;
	    image: string;
	    imageId: string;
	    state: string;
	    status: string;
	    // Go type: time
	    created: any;
	    sizeRw: number;
	    labels: Record<string, string>;
	    volumes: string[];
	    networkIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new Container(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.image = source["image"];
	        this.imageId = source["imageId"];
	        this.state = source["state"];
	        this.status = source["status"];
	        this.created = this.convertValues(source["created"], null);
	        this.sizeRw = source["sizeRw"];
	        this.labels = source["labels"];
	        this.volumes = source["volumes"];
	        this.networkIds = source["networkIds"];
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
	export class EngineInfo {
	    name: string;
	    version: string;
	    apiVersion: string;
	    os: string;
	    osType: string;
	    architecture: string;
	    kernelVersion: string;
	    storageDriver: string;
	    rootDir: string;
	    cpus: number;
	    memoryBytes: number;
	
	    static createFrom(source: any = {}) {
	        return new EngineInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.apiVersion = source["apiVersion"];
	        this.os = source["os"];
	        this.osType = source["osType"];
	        this.architecture = source["architecture"];
	        this.kernelVersion = source["kernelVersion"];
	        this.storageDriver = source["storageDriver"];
	        this.rootDir = source["rootDir"];
	        this.cpus = source["cpus"];
	        this.memoryBytes = source["memoryBytes"];
	    }
	}
	export class Image {
	    id: string;
	    tags: string[];
	    // Go type: time
	    created: any;
	    size: number;
	    sharedSize: number;
	    labels: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Image(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.tags = source["tags"];
	        this.created = this.convertValues(source["created"], null);
	        this.size = source["size"];
	        this.sharedSize = source["sharedSize"];
	        this.labels = source["labels"];
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
	export class Network {
	    id: string;
	    name: string;
	    driver: string;
	    scope: string;
	    ingress: boolean;
	    // Go type: time
	    created: any;
	    labels: Record<string, string>;
	    internal: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Network(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.driver = source["driver"];
	        this.scope = source["scope"];
	        this.ingress = source["ingress"];
	        this.created = this.convertValues(source["created"], null);
	        this.labels = source["labels"];
	        this.internal = source["internal"];
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
	export class Volume {
	    name: string;
	    driver: string;
	    // Go type: time
	    created: any;
	    size: number;
	    labels: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Volume(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.driver = source["driver"];
	        this.created = this.convertValues(source["created"], null);
	        this.size = source["size"];
	        this.labels = source["labels"];
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
	export class Inventory {
	    // Go type: time
	    takenAt: any;
	    engine: EngineInfo;
	    containers: Container[];
	    images: Image[];
	    volumes: Volume[];
	    networks: Network[];
	    buildCache: BuildCacheRecord[];
	
	    static createFrom(source: any = {}) {
	        return new Inventory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.takenAt = this.convertValues(source["takenAt"], null);
	        this.engine = this.convertValues(source["engine"], EngineInfo);
	        this.containers = this.convertValues(source["containers"], Container);
	        this.images = this.convertValues(source["images"], Image);
	        this.volumes = this.convertValues(source["volumes"], Volume);
	        this.networks = this.convertValues(source["networks"], Network);
	        this.buildCache = this.convertValues(source["buildCache"], BuildCacheRecord);
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

export namespace main {
	
	export class ScanResult {
	    inventory: inventory.Inventory;
	    report: cleanup.Report;
	    safeSize: number;
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inventory = this.convertValues(source["inventory"], inventory.Inventory);
	        this.report = this.convertValues(source["report"], cleanup.Report);
	        this.safeSize = source["safeSize"];
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

export namespace sweep {
	
	export class Request {
	    kind: string;
	    id: string;
	    verdict: string;
	
	    static createFrom(source: any = {}) {
	        return new Request(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.id = source["id"];
	        this.verdict = source["verdict"];
	    }
	}
	export class Result {
	    kind: string;
	    id: string;
	    name: string;
	    status: string;
	    message: string;
	    freed: number;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.id = source["id"];
	        this.name = source["name"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.freed = source["freed"];
	    }
	}

}

