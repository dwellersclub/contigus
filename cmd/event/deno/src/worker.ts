import { EventReference } from "./event.ts";
import { logger } from "./logger.ts";

export class Listener {
  id?: string;
  glob?: string;
}

export class BaseWorker {
  private id: string;
  private listeners: Map<string, Listener>;

  constructor(id: string){
    this.id = id;
    this.listeners = new Map<string, Listener>();
  }

  compile(filePath: string): string {
    return ""
  }

  run(id: string): void {
    console.log('listening')
  }

  match(event: EventReference): boolean {
    logger.info("matching", event)
    return true
  }

  process(event: EventReference): void {
    logger.info("process", event)
    return
  }

  addListener(glob: string): string {
    return ""
  }

  clearListeners(): void {
    this.listeners.clear()
  }

}