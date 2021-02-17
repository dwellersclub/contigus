
import { logger } from "./logger.ts";
import { EventHandler } from "./event.ts";
import { Consumer } from "./message.ts";

const handler = (e: Event) => {
    logger.info(`got ${e.type} event in event handler (main)`)
};

window.addEventListener("load", handler);

window.addEventListener("unload", handler);

window.onload = (e: Event): void => {

    const queueName = "events"
    const subject = "*"

    logger.info(`got ${e.type} event in onload function (main)`);

    const workers = new Worker(new URL("./workers.ts", import.meta.url).href, { type: "module", deno: true });
    workers.postMessage({ action: "start", args: {host: "localhost", port: "8888"} });
    
    const onEvent:EventHandler = (event) => {
        workers.postMessage({ action: "new_event", args: {event}});
    }

    const eventConsumer = new Consumer(queueName, "localhost:4222")
    eventConsumer.listen(subject, onEvent)
    
};

window.onunload = (e: Event): void => {
    logger.info(`got ${e.type} event in onunload function (main)`);
};

logger.info("log from main script");