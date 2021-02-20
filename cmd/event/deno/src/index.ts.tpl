import { uuid } from "https://deno.land/x/uuid/mod.ts";
import { setUp } from "./[eventId].js";

const globs = {}
const globConfigs = {}
const projects = {}

const events = {
    "on": (glob, handler) => {
        const myUUID = uuid();
        globs[myUUID] = {
            glob,
            handler,
            id: myUUID
        }
        globConfigs[myUUID] = glob
        }
}
const env = {}
const jobs = {}

self.onmessage = async (e) => {
    const {action, args} = e.data

    switch(action) {
        case "start":
            start()
        break
        case "handle_event":
            const { event } = args
            args.matches.forEach((id) => {
                const globConfig = globs[id]
                if(globConfig && globConfig.handler) {
                    globConfig.handler(event)
                }
            })
        break
    }

};

function start() {
    try {
        setUp({projects, events, env, jobs})
        self.postMessage({ action: "register", args: {globConfigs} })
    } catch(err) {
        console.error(err)
    }
}