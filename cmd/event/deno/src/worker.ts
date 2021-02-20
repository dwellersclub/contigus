import { EventReference } from "./event.ts";
import minimatch from 'https://deno.land/x/minimatch/index.js';
import { logger } from "./logger.ts";

export class Listener {
  id: string;
  matcher;

  constructor(id: string, glob?: string) {
    this.id = id
    this.matcher = minimatch.Minimatch(glob, { debug: false })
    this.match = this.match.bind(this);
  }

  match(glob?: string): boolean {
    return this.matcher.match(glob)
  }

}

export class BaseWorker {
  private id: string;
  private filePath: string;
  private listeners: Map<string, Listener>;
  private worker?: Worker;

  constructor(id: string, filePath: string){
    this.id = id;
    this.filePath = filePath;
    this.listeners = new Map<string, Listener>();
    this.run = this.run.bind(this);
    this.addListener = this.addListener.bind(this);
    this.terminate = this.terminate.bind(this);
    this.clearListeners = this.clearListeners.bind(this);

    logger.info(`new worker [${id}] with source [${filePath}]`)
  }

  run(): void {

    this.terminate()
    this.clearListeners()

    const webWorker = new Worker(new URL(this.filePath, import.meta.url).href, { type: "module", deno: true });
    webWorker.onerror = (e: ErrorEvent) => { 
      e.preventDefault()
      console.error(e.message) 
    }

    webWorker.onmessage = (e) => {
      if(e.data.action == "register"){
        const globConfigs = e.data.args['globConfigs']
        Object.keys(globConfigs).forEach(key => {
          this.addListener(key, globConfigs[key])
        });
      }
    }
    webWorker.postMessage({ action: "start", args: {} })
    this.worker = webWorker
  }

  terminate(): void {
    if (this.worker) {
      logger.info(`terminate worker [${this.id}]`)
      this.worker.terminate()
      this.worker = undefined
    }
  }

  match(event: EventReference): string[] {
    const matches: string[] = []

    this.listeners.forEach((listener: Listener) => {
      if (listener && listener.match(event.source)) {
        matches.push(listener.id)
      }
    });

    logger.info("matches", matches);

    return matches
  }

  process(event: EventReference, matches: string[]): void {
    logger.info("process", event)
    if (this.worker) {
      this.worker.postMessage({action: "handle_event", args : {event, matches}})
    }
    return
  }

  addListener(id: string, glob?: string): void {
    logger.info("register", {id, glob})
    this.listeners.set(id, new Listener(id, glob))
  }

  clearListeners(): void {
    this.listeners.clear()
  }

}