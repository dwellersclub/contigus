import { serve, Server } from "https://deno.land/std/http/server.ts";
import {
  acceptWebSocket,
  isWebSocketCloseEvent,
  isWebSocketPingEvent,
  WebSocket,
} from "https://deno.land/std/ws/mod.ts";

import { EventReference } from "./event.ts";
import { BaseWorker } from "./worker.ts";
import { logger } from "./logger.ts";

export interface BaseWorkers {
  dispatch(event: EventReference): void;
  getWorkers(): BaseWorker[];
  install(path: string): BaseWorker;
  uninstall(id: string): boolean;
}

export class Workers {
  private host: string;
  private port: string;
  private password: string;
  private server: Server;

  constructor(host: string, port: string) {
    this.password = randomChars(35);
    this.host = host;
    this.port = port;
    this.server = serve(`${this.host}:${this.port}`)
  }

  async start(): Promise<void> {
    logger.info(`starting websocket server on ${this.host}:${this.port}`);
    for await (const req of this.server) {
      const { conn, r: bufReader, w: bufWriter, headers } = req;
      acceptWebSocket({
        conn,
        bufReader,
        bufWriter,
        headers,
      })
        .then(this.handleWs.bind(this))
        .catch(async (err) => {
          logger.error(`failed to accept websocket: ${err}`);
          await req.respond({ status: 400 });
        });
    }
  }

  async stop(): Promise<void> {
    logger.info(`starting websocket client to ${this.host}:${this.port}`);
    const endpoint = `ws://${this.host}:${this.port}`;
    try {
      const webSocket = new WebSocket(endpoint);
      webSocket.addEventListener('open', () => webSocket.send(`close|${this.password}`));
    } catch (err) {
      logger.error(`failed to connect to websocket: ${err}`);
    }
  }

  async handleWs(sock: WebSocket): Promise<void> {
    logger.info("socket connected!");
    try {
      for await (const ev of sock) {
        if (typeof ev === "string") {
          // text message
          logger.info("ws:Text", ev);
          await sock.send(ev);
        } else if (ev instanceof Uint8Array) {
          // binary message
          logger.info("ws:Binary", ev);
        } else if (isWebSocketPingEvent(ev)) {
          const [, body] = ev;
          // ping
          logger.info("ws:Ping", body);
        } else if (isWebSocketCloseEvent(ev)) {
          // close
          const { code, reason } = ev;
          logger.info("ws:Close", code, reason);
        }
      }
    } catch (err) {
      logger.error(`failed to receive frame: ${err}`);
      if (!sock.isClosed) {
        await sock.close(1000).catch(console.error);
      }
    }
  }
}

const randomChars = (n: number) =>
  Array(n).fill(0).map((elt: number) => {
    return Math.ceil(Math.random() * 35 + elt).toString(36);
  }).join("");

let workers: Workers = new Workers("", "")

self.onmessage = async (e: MessageEvent) => {
  switch(e.data.action) {
    case 'start':
      workers = new Workers(e.data.args.host, e.data.args.port)
      workers.start()
      break
    case 'stop':
      if(workers){
        workers.stop()
        setTimeout(() => self.close(), 1000)
      }
      break
    case 'new_event':
      logger.info(`process event ${e.data.args.event}`)
      //check concatenation of {subject}.{eventType} to dispatch to corresponding worker
      break
  }
};