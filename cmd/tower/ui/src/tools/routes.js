import UniversalRouter from 'universal-router'
import {h, Component} from 'preact'

const NotFoundUI = (pathname) => <h1>Page '${pathname}' Not Found</h1>
const ErrorUI = <h1>Oops! Something went wrong</h1>

export const createRoutes = (onPathChanged, isAnonymous, routers = [], config = {}, actions) => {

  if (!Array.isArray(routers)) {
    throw new TypeError('Invalid routes')
  }

  const loginPath = config.loginPath || '/signin'
  const errorHandler = config.errorHandler || ((error, context) => {
    console.error(error, context)
    return error.status === 404 ? NotFoundUI(context.pathname) : ErrorUI
  })

  const options = {
    resolveRoute(context, params) {

      if (typeof context.route.action === 'function') {

        let secured = false

        if (context.route.hasOwnProperty('secured')) {
          secured = context.route.secured
        }

        if (secured && isAnonymous()) {
          for (let index = 0; index < context.router.root.children.length; index++) {
            const element = context.router.root.children[index]
            if (element.path === loginPath) {
              const result = element.action(context, params)
              if (onPathChanged) { onPathChanged(element.path) }
              return result
            }
          }
          return undefined
        }
        const result = context.route.action(context, params)
        if (onPathChanged) { onPathChanged(context.path) }
        return result
      }
      return undefined
    },
    errorHandler
  }

  const router = new UniversalRouter([], options)

  const routes = {}

  const addRoutes = (routes) => BuildRoutes(routes).forEach((item) => {
    console.log('Adding route ', item.path)
    router.root.children.push(item)
  })

  routers.forEach((setupHandler) =>  setupHandler({addRoutes}))

  addRoutes(routes)

  return  router
}

function BuildRoutes(routes) {
  const appRoutes = []
  Object.keys(routes).forEach((key) => {

    let actionPath = ''
    const pathConfig = key.split('|')
    let secured = true

    if (pathConfig.length === 1) {
      actionPath = pathConfig[0]
    }

    if (pathConfig.length > 1) {
      secured = false
      actionPath = pathConfig[1]
    }

    const action = routes[key]
    if (typeof action == 'function')  {
      appRoutes.push({path: actionPath, action, secured})
    } else {
      appRoutes.push({path: actionPath,children: BuildRoutes(action), action: () => {}, secured})
    }
  })
  return appRoutes
}
