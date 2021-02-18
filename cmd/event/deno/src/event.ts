
export class EventReference {
    id?: string
    source?: string
    type?: string
}

export type EventHandler = (evt: string) => void;