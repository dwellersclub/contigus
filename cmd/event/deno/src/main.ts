
import { logger } from "./logger.ts";
import { EventHandler } from "./event.ts";
import { Consumer } from "./message.ts";

try {
    
const workers = new Worker(new URL("./workers.ts", import.meta.url).href, { type: "module", deno: true });
workers.onerror = (e: ErrorEvent) => { 
    e.preventDefault()
    console.error(e.message) 
}

const handler = (e: Event) => {
    logger.info(`got ${e.type} event in event handler (main)`)
};

window.addEventListener("load", handler);

window.addEventListener("unload", handler);

window.onload = (e: Event): void => {
    // should come from envs
    const queueName = "events"
    const subject = "*"
    const server = "localhost:4222"
    const webSocketServer = "localhost"
    const webSocketPort = "8888"

    logger.info(`got ${e.type} event in onload function (main)`);

    workers.postMessage({ action: "start", args: {host: webSocketServer, port: webSocketPort} });
    
    const onEvent:EventHandler = (event) => {
        workers.postMessage({ action: "new_event", args: {event}});
    }

    const eventConsumer = new Consumer(queueName, server)
    eventConsumer.listen(subject, onEvent)
    
};

window.onunload = (e: Event): void => {
    logger.info(`got ${e.type} event in onunload function (main)`);
};

logger.info("log from main script");
} catch (error) {
    logger.error(error)
}