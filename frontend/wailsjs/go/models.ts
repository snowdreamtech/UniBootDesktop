export namespace config {
	
	export class AppConfig {
	    mode: string;
	    autoCheckUpdate: boolean;
	    theme: string;
	    githubProxy: string;
	    fileSystem: string;
	    proxyProtocol: string;
	    proxyHost: string;
	    proxyPort: number;
	    proxyUser: string;
	    proxyPassword: string;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.autoCheckUpdate = source["autoCheckUpdate"];
	        this.theme = source["theme"];
	        this.githubProxy = source["githubProxy"];
	        this.fileSystem = source["fileSystem"];
	        this.proxyProtocol = source["proxyProtocol"];
	        this.proxyHost = source["proxyHost"];
	        this.proxyPort = source["proxyPort"];
	        this.proxyUser = source["proxyUser"];
	        this.proxyPassword = source["proxyPassword"];
	    }
	}

}

export namespace disk {
	
	export class DiskInfo {
	    device: string;
	    name: string;
	    size: number;
	    formatted: string;
	    freeSpace: number;
	    freeFormatted: string;
	    isRemovable: boolean;
	    isSystem: boolean;
	    usbVersion: string;
	    usbSpeed: string;
	    vendor: string;
	    fileSystem: string;
	    partitionScheme: string;
	    writable: boolean;
	    serialNumber: string;
	    vendorId: string;
	    productId: string;
	    smartStatus: string;
	    busPower: string;
	    busPowerUsed: string;
	    sectorSize: string;
	    transportProtocol: string;
	    bootStatus: string;
	    controllerVendor: string;
	    isFakeUsb3: boolean;
	    protocolCode: string;
	
	    static createFrom(source: any = {}) {
	        return new DiskInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.device = source["device"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.formatted = source["formatted"];
	        this.freeSpace = source["freeSpace"];
	        this.freeFormatted = source["freeFormatted"];
	        this.isRemovable = source["isRemovable"];
	        this.isSystem = source["isSystem"];
	        this.usbVersion = source["usbVersion"];
	        this.usbSpeed = source["usbSpeed"];
	        this.vendor = source["vendor"];
	        this.fileSystem = source["fileSystem"];
	        this.partitionScheme = source["partitionScheme"];
	        this.writable = source["writable"];
	        this.serialNumber = source["serialNumber"];
	        this.vendorId = source["vendorId"];
	        this.productId = source["productId"];
	        this.smartStatus = source["smartStatus"];
	        this.busPower = source["busPower"];
	        this.busPowerUsed = source["busPowerUsed"];
	        this.sectorSize = source["sectorSize"];
	        this.transportProtocol = source["transportProtocol"];
	        this.bootStatus = source["bootStatus"];
	        this.controllerVendor = source["controllerVendor"];
	        this.isFakeUsb3 = source["isFakeUsb3"];
	        this.protocolCode = source["protocolCode"];
	    }
	}

}

export namespace firmware {
	
	export class FirmwareMapping {
	    releaseName: string;
	    targetPath: string;
	    description: string;
	    isReserved: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FirmwareMapping(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.releaseName = source["releaseName"];
	        this.targetPath = source["targetPath"];
	        this.description = source["description"];
	        this.isReserved = source["isReserved"];
	    }
	}
	export class UniBootReleaseAsset {
	    name: string;
	    browser_download_url: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new UniBootReleaseAsset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.browser_download_url = source["browser_download_url"];
	        this.size = source["size"];
	    }
	}
	export class UniBootReleaseInfo {
	    tagName: string;
	    name: string;
	    publishedAt: string;
	    body: string;
	    assets: UniBootReleaseAsset[];
	    localTag: string;
	    hasUpdate: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UniBootReleaseInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tagName = source["tagName"];
	        this.name = source["name"];
	        this.publishedAt = source["publishedAt"];
	        this.body = source["body"];
	        this.assets = this.convertValues(source["assets"], UniBootReleaseAsset);
	        this.localTag = source["localTag"];
	        this.hasUpdate = source["hasUpdate"];
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

export namespace installer {
	
	export class DeployResult {
	    success: boolean;
	    mode: string;
	    target: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new DeployResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.mode = source["mode"];
	        this.target = source["target"];
	        this.message = source["message"];
	    }
	}

}

export namespace qemu {
	
	export class QEMUStatus {
	    installed: boolean;
	    path: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new QEMUStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.path = source["path"];
	        this.version = source["version"];
	    }
	}

}

export namespace updater {
	
	export class UpdateStatus {
	    hasUpdate: boolean;
	    currentTag: string;
	    latestTag: string;
	    downloadUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasUpdate = source["hasUpdate"];
	        this.currentTag = source["currentTag"];
	        this.latestTag = source["latestTag"];
	        this.downloadUrl = source["downloadUrl"];
	    }
	}

}

