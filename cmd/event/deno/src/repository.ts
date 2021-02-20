
import { EventReference } from "./event.ts";

const compileConfig = {
  "id": "123456",
  "url" : "/Users/stanpanza/opt/project/go/src/github.com/dwellersclub/contigus/cmd/event/deno/example.js"
}

class EventRepository {
    get(ref: EventReference): Promise<string> {
        return new Promise<string>((resolve, reject) => {
            setTimeout(() => {
              resolve(JSON.stringify(compileConfig));
            }, 300);
          });
    }
}

// call the API to retieve the event
// TODO handle authentication to be able to access restricted events

export const Events = new EventRepository()
