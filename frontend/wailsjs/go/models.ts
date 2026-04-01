export namespace main {
	
	export class AppConfig {
	    upperLimit: number;
	    lowerLimit: number;
	    intervalMinutes: number;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.upperLimit = source["upperLimit"];
	        this.lowerLimit = source["lowerLimit"];
	        this.intervalMinutes = source["intervalMinutes"];
	    }
	}

}

