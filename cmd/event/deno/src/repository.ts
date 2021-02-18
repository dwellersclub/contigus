
import { EventReference } from "./event.ts";

export interface Events {
    get(ref: EventReference): Promise<string>;
}

// call the API to retieve the event
// TODO handle authentication to be able to access restricted events