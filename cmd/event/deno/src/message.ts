import {
  connect,
  ConnectionOptions,
  StringCodec,
  Subscription,
} from "https://raw.githubusercontent.com/nats-io/nats.deno/main/src/mod.ts";
import { EventHandler } from "./event.ts";
import { logger } from "./logger.ts";

const sc = StringCodec();

export class Consumer {
  private queueName: string;
  private server: string;

  constructor(queueName: string, server: string) {
    this.queueName = queueName;
    this.server = server;
  }

  async listen(key: string, onEvent: EventHandler): Promise<void> {
    try {
      const opts = { servers: this.server, debug: true } as ConnectionOptions;
      const nc = await connect(opts);
      const sub = nc.subscribe(key, { queue: this.queueName });
      this.handleRequest(this.queueName, sub, onEvent);
    } catch (err) {
      logger.error("error connecting", err);
    }
  }

  async handleRequest(name: string, s: Subscription, onEvent: EventHandler): Promise<void> {
    const p = 12 - name.length;
    const pad = "".padEnd(p);
    for await (const m of s) {
      // respond returns true if the message had a reply subject, thus it could respond
      onEvent(sc.decode(m.data))
      if (m.respond(m.data)) {
        logger.info(
          `[${name}]:${pad} #${s.getProcessed()} echoed ${sc.decode(m.data)}`,
        );
      } else {
        logger.info(
          `[${name}]:${pad} #${s.getProcessed()} ignoring request - no reply subject`,
        );
      }
    }
  }
}
