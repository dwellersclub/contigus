import produce, { applyPatches } from 'immer'
import Trie from './trie'

const get = (p, o) =>
    p.reduce((xs, x) => (xs && xs.hasOwnProperty(x) ? xs[x] : null), o)

class BaseStore {
    constructor(defaultState) {
        this.actionLocator = null
        this.time = new Date()
        this.state = defaultState == null ? {} : Object.assign({}, defaultState)
        this.trie = new Trie()

        this.update = this.update.bind(this)
        this.watch = this.watch.bind(this)
        this.watchKeys = this.watchKeys.bind(this)
        this.getPath = this.getPath.bind(this)
        this.initState = this.initState.bind(this)
        this.clearWatch = this.clearWatch.bind(this)
        this.setActionLocator = this.setActionLocator.bind(this)
        this.mount = this.mount.bind(this)
    }

    initState(defaultState) {
        this.state = Object.assign({}, defaultState)
    }

    addState(namespace, defaultState) {
        this.state[namespace] = Object.assign({}, defaultState)
    }

    setActionLocator(actionLocator) {
        this.actionLocator = actionLocator
    }

    clearWatch() {
        this.trie.clear()
    }

    getValue(path) {
        return get(path.split('/'), this.state)
    }

    update(updateHandler) {
        const changes = []
        const revertChanges = []

        const changesHandler = (patches, revert) => {
            changes.push(...patches)
            revertChanges.push(...revert)
        }

        produce(this.state, (draft) => updateHandler(draft), changesHandler)

        this.state = applyPatches(this.state, changes)

        const handlers = {}

        changes.forEach((item) => {
            let updatepath = ``
            item.path.forEach((path) => {
                updatepath = `${updatepath}/${encodeURIComponent(path)}`
            })

            const results = this.trie.all(`${updatepath}/*`)

            if (results) {
                results.forEach((result) => {
                    result.values().forEach((value) => {
                        const id = value.id
                        let handlerConfig = handlers[id]
                        if (!handlerConfig) {
                            handlerConfig = []
                            handlers[id] = handlerConfig
                        }

                        let values = {}
                        if (value.mapping.hasOwnProperty('mappingKey')) {
                            const data = get(value.mapping.path, this.state)

                            if (value.mapping.mappingKey === '') {
                                values = { ...values, ...data }
                            } else {
                                values[value.mapping.mappingKey] = data
                            }
                        }

                        handlers[id].push({ listener: value.listener, values })
                    })
                })
            }
        })

        Object.keys(handlers).forEach((key) => {
            try {
                const listener = handlers[key][0].listener
                let values = {}
                handlers[key].forEach((item) => {
                    values = { ...values, ...item.values }
                })
                listener(values)
            } catch (error) {
                console.error(error)
            }
        })

        return { changes, revertChanges }
    }

    watchKeys(key, listener) {
        const keys = []
        if (Array.isArray(key)) {
            keys.push(...key)
        } else {
            keys.push(key)
        }

        const id = Math.floor(Math.random() * 100000 + 1)

        keys.forEach((item) => {
            const mapping = this.getPath(item)
            this.trie.add(mapping.key, { id, mapping, listener })
        })

        return id
    }

    getPath(rawKey) {
        const config = rawKey.split(':')

        let mappingKey = ''
        let path = ''
        let key = rawKey

        if (config.length == 2) {
            mappingKey = config[0]
            key = config[1]
            path = key
        } else if (config.length == 3) {
            mappingKey = config[0]
            path = config[1]
            key = config[2]
        }

        if (key && key.startsWith('/')) {
            const index = key.indexOf('/*')
            if (index > -1) {
                path = key.slice(0, index)
            }
        }
        return { mappingKey, path: path.slice(1).split('/'), key }
    }

    watch(component, fn, defaultMappers, onlyChanged) {
        const mapping = []
        const mappingConfig = []

        let mappers = []

        const sources = [defaultMappers, component.dataMapper]

        if (component.prototype) {
            sources.push(component.prototype.dataMapper)
        }

        sources.forEach((item) => {
            if (item) {
                mappers = mappers.concat(item)
            }
        })

        mappers.forEach((key) => {
            mapping.push(key)
            mappingConfig.push(this.getPath(key))
        })

        const getValues = () => {
            let values = {}
            mappingConfig.forEach((value) => {
                if (value.hasOwnProperty('mappingKey')) {
                    if (value.mappingKey === '') {
                        values = { ...values, ...get(value.path, this.state) }
                    } else {
                        values[value.mappingKey] = get(value.path, this.state)
                    }
                }
            })
            return values
        }

        const id = this.watchKeys(mapping, (values) => {
            if (!onlyChanged) {
                values = Object.assign(values, getValues())
            }
            return fn(values)
        })

        const close = () =>
            mappingConfig.forEach((value) =>
                this.trie.deleteById(value.key, id)
            )

        return { close, getValues }
    }

    mount(
        targetClass,
        defaultMapper,
        cleanUpHandler = () => {},
        onlyChanged = false
    ) {
        const functs = {}
        if (
            targetClass &&
            targetClass.constructor &&
            targetClass.hasOwnProperty('stateTypes')
        ) {
            Object.keys(targetClass.stateTypes).forEach((key) => {
                if (key.startsWith('$')) {
                    const name = key.slice(1)
                    if (this.actionLocator) {
                        const funct = this.actionLocator(name)
                        if (funct) {
                            functs[key] = funct
                        }
                    }
                }
            })
        }

        let stateUpdater = () => {}
        let setUpdater = (updater) => {
            stateUpdater = updater
        }

        const storeWatcher = this.watch(
            targetClass,
            (values) => stateUpdater(Object.assign(functs, values)),
            defaultMapper,
            onlyChanged
        )

        const values = Object.assign(functs, storeWatcher.getValues())
        const unmount = () => {
            storeWatcher.close()
            cleanUpHandler()
        }

        return { values, unmount, setUpdater }
    }
}

const baseStore = new BaseStore({})
export const store = baseStore
