import {h, render} from 'preact'
import { Application } from './page/app.js';

import {
  store, buildActions,
  createRoutes, web
} from './tools'

import { getDefaultState } from './state'
import { actions, actionReducers } from './task'
import { routes } from './routes'

const defaultLanguage = 'en_GB'

store.initState(getDefaultState(defaultLanguage))

let anonymous = true
let logDebug = true

const handlers = buildActions(store, actions, actionReducers, () => logDebug, {}, console)
const routeChangeHandler = (location) => console.log(location)
const router = createRoutes(routeChangeHandler, () => { return anonymous }, routes, {}, handlers)

window.childNode = null

store.watchKeys('location:/location/*', (values) => {
  router.resolve({ pathname: values.location.currentPath }).then(childNode => {
    store.update((state) => {
      web.setCookie('ui_location', values.location.currentPath)
      state.ui.errMsg = null
      window.childNode = childNode
      state.ui.childNode = new Date().getTime()
    })
  })
})

store.watchKeys('user:/user/*', (values) => {
  if (anonymous && values.user.username !== null && values.user.username !== "visitor") {
    anonymous = false
  } else if (!anonymous && values.user.username === "visitor") {
    anonymous = true
    store.initState(getDefaultState(defaultLanguage))
  }
})

render( <Application />, document.body, null);

setTimeout(() => { 
  handlers.go('/')
}, 300)