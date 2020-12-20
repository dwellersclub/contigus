import produce, { applyPatches } from 'immer'
import Trie from './trie'

const get = (p, o) =>
  p.reduce((xs, x) => (xs && xs.hasOwnProperty(x) ? xs[x] : null), o)

class BaseStore {
  constructor(defaultState) {
    this.actionLocator = null
    this.time = new Date()
    this.state =
      defaultState == null ? {} : Object.assign({}, defaultState)
    this.trie = new Trie()

    this.update = this.update.bind(this)
    this.watch = this.watch.bind(this)
    this.watchKeys = this.watchKeys.bind(this)
    this.getPath = this.getPath.bind(this)
    this.initState = this.initState.bind(this)
    this.clearWatch = this.clearWatch.bind(this)
    this.setActionLocator = this.setActionLocator.bind(this)
  }

  initState(defaultState) {
    this.state = Object.assign({}, defaultState)
  }

  setActionLocator(actionLocator) {
    this.actionLocator =  actionLocator
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

    produce(
      this.state,
      (draft) => updateHandler(draft),
      changesHandler
    )

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

            const values = {}
            if (value.mapping.mappingKey) {
              values[value.mapping.mappingKey] = get(
                value.mapping.path,
                this.state
              )
            }
            handlers[id].push({ listener: value.listener, values })
          })
        })
      }
    })

    Object.keys(handlers).forEach((key) => {
      try {
        const listener = handlers[key][0].listener
        const values = {}
        handlers[key].forEach((item) => {
          Object.assign(values, item.values)
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

    const sources = [
      defaultMappers,
      component.dataMapper
    ]

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
      const values = {}
      mappingConfig.forEach((value) => {
        if (value.mappingKey) {
          values[value.mappingKey] = get(value.path, this.state)
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

    const close = () => mappingConfig.forEach((value) => this.trie.deleteById(value.key, id))

    return { close, getValues }
  }

  mount(target, setStoreProps, defaultMapper, initHandler = () => { }, cleanUpHandler = () => { }, onlyChanged = false) {

    const functs = {}
    if (target && target.hasOwnProperty('propTypes')) {
      Object.keys(target.propTypes).forEach((key) => {
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

    return () => {
      const storeWatcher = this.watch(
        target,
        (values) => setStoreProps(Object.assign(functs, values), false),
        defaultMapper,
        onlyChanged
      )
      setStoreProps(Object.assign(functs,storeWatcher.getValues()), true)
      if (initHandler) {
        initHandler()
      }
      return () => {
        storeWatcher.close()
        cleanUpHandler()
      }
    }
  }
}

const baseStore = new BaseStore({})
export const store = baseStore
