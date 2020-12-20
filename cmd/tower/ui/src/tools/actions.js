const repeat = (str, times) => new Array(times + 1).join(str)

const pad = (num, maxLength) =>
  repeat('0', maxLength - num.toString().length) + num

const formatTime = (time) =>
  `${pad(time.getHours(), 2)}:${pad(time.getMinutes(), 2)}:${pad(
    time.getSeconds(),
    2
  )}.${pad(time.getMilliseconds(), 3)}`

const lightDisplay = 'color: gray font-weight: lighter'
const boldDisplay = 'color: black font-weight: bold'

const displayChanges = (changes) =>
  changes.reduce((result, item) => {
    result.push([item.path.join('.'), item.value])
    return result
  }, [])

export const buildActions = (store, tasks, reducers, isDebug) => {
  const actions = {}

  const defaultReducer = () => null

  const resultHandler = function(
    handler,
    key,
    actionArgs,
    completionHandler
  ) {
    return (...results) => {
      try {
        const context = { actions }
        const stateHandler = handler(...results, context, actionArgs)
        if (stateHandler) {
          let stateCompletionHandler = null
          const diff = store.update((state) => {
            stateCompletionHandler = stateHandler(state)
          })

          if (isDebug()) {
            console.groupCollapsed(
              completionHandler(),
              lightDisplay,
              lightDisplay,
              boldDisplay
            )
            console.debug('results', results)
            if (diff.changes.length > 0) {
              console.debug(
                'state',
                displayChanges(diff.changes)
              )
            }
            console.groupEnd()
          } else {
            completionHandler()
          }

          if (stateCompletionHandler instanceof Function) {
            setTimeout(stateCompletionHandler, 100)
          }
        } else if (isDebug()) {
          console.debug(
            completionHandler(),
            lightDisplay,
            lightDisplay,
            boldDisplay,
            results
          )
        } else {
          completionHandler()
        }
      } catch (err) {
        console.error(
          completionHandler(),
          lightDisplay,
          lightDisplay,
          boldDisplay,
          err.message
        )
      }
    }
  }

  const perfMark = (actionName, funcArgs) => {
    if (actionName.includes('login') || actionName.includes('register')
      || actionName.includes('mergeState')) {
      funcArgs.toJSON = () => funcArgs.map(() => 'XXXXXXXXX')
    }

    const start = new Date()
    return () => `%c @ ${formatTime(start)} %c[${new Date().getTime() -
      start.getTime()} ms] | %c${actionName} ${JSON.stringify(
      funcArgs
    )} `
  }

  tasks.names().forEach((key) => {
    const handler = reducers[key] ? reducers[key] : defaultReducer
    const obs = tasks.action(key)
    if (!obs) {
      console.error(key, 'not found')
      return
    }

    actions[key] = (...actionArgs) => {
      const endPerf = perfMark(key, actionArgs.slice())

      const result = obs(...actionArgs)

      return Promise.resolve(result).then((value) => {
        resultHandler(handler, key, actionArgs, endPerf)(value)
      })
    }
  })

  const actionKeys = Object.keys(actions)

  actions.batch = function() {
    let args = []
    let handlers = []
    const delays = {}
    let cancel = false
    let label = ''

    const functions = {
      run(sequential) {

        let endPerfLogs = new Array(args.length)
        let promisesToExec = new Array(args.length)

        const resultPromiseHandler = (values) => {
          if (!cancel) {
            const context = { actions }
            const stateHandlers = []
            handlers.forEach((handler, i) => {
              const stateHandler = handler(
                ...[values[i]],
                context,
                args[i].actionArgs
              )
              stateHandlers.push(stateHandler)
            })

            const stateHandlerResults = []
            const diff = store.update((state) => {
              stateHandlers.forEach((stateHandler, i) => {
                stateHandlerResults.push(
                  stateHandler instanceof Function ? stateHandler(state) : null
                )
              })
            })

            if (isDebug()) {
              console.group(label)
              endPerfLogs.forEach((endPerfLog, i) => {
                console.groupCollapsed(
                  endPerfLog,
                  lightDisplay,
                  lightDisplay,
                  boldDisplay
                )
                console.debug('results', values[i])
                console.groupEnd()
              })

              if (diff.changes.length > 0) {
                console.debug(
                  'state',
                  displayChanges(diff.changes)
                )
              }
              console.groupEnd()
            }

            stateHandlerResults.forEach((stateHandlerResult) => {
              if (stateHandlerResult instanceof Function) {
                setTimeout(stateHandlerResult, 50)
              }
            })
          }

          args = null
          endPerfLogs = null
          handlers = null
          promisesToExec = null
        }

        label = `${args.length} batched action(s) ${sequential ? 'sequential' : ''}`
        args.forEach((arg, index) => {
          const key = arg.action
          const handler = reducers[key]
            ? reducers[key]
            : defaultReducer
          handlers.push(handler)
          const obs = tasks.action(key)

          if (!obs) {
            console.error(key, 'not found')
            return
          }

          const endPerf = perfMark(key, arg.actionArgs.slice())

          if (obs instanceof Function) {
            promisesToExec[index] = {
              key,
              action: (delay) => new Promise((resolve, reject) => {

                const trigger = (result) => Promise.resolve(result).then((item) => {
                  resolve(item)
                  endPerfLogs[index] = endPerf()
                }).catch(reject)

                const handler = () => trigger(obs(...arg.actionArgs))

                if (delay) {
                  setTimeout(handler, delay)
                } else {
                  handler()
                }

              }),
              delay: delays[key]
            }
          } else {
            console.error(` invalid method for action ${key}`)
          }
        })

        if (sequential) {
          return promisesToExec.reduce(
            (promise, execConfig) => promise.then((result) =>
              execConfig.
                action(execConfig.delay).
                catch(console.error).
                then((value) => {
                  result.push(value)
                  return result
                })),
            Promise.resolve([])
          ).then(resultPromiseHandler).catch(console.error)
        }

        const promises = []
        for (const execConfig of promisesToExec) {
          promises.push(execConfig.action(execConfig.delay))
        }

        return Promise.all(promises)
          .then(resultPromiseHandler)
          .catch(console.error)
      },
      cancel() {
        cancel = true
      }
    }

    actionKeys.forEach((action) => {
      functions[action] = (...actionArgs) => {
        args.push({ action, actionArgs })
        let localFunctions = {}
        localFunctions = {
          ...functions,
          delay: (second) => {
            delays[action] = second
            return localFunctions
          }
        }
        return localFunctions
      }
    })

    return functions
  }

  store.setActionLocator((name) => {
    if (actions.hasOwnProperty(name)) {
      return actions[name]
    } else {
      console.error(`action '${name}' doesn't exist`)
    }
  })

  return actions
}
