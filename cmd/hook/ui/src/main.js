import {html, render} from 'lit-html';
import './page/app.js';

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

store.watchKeys('location:/location/*', (values) => {
  router.resolve({ pathname: values.location.currentPath }).then(childNode => {
    store.update((state) => {
      web.setCookie('ui_location', values.location.currentPath)
      state.ui.errMsg = null
      Object.assign(state.ui, { childNode })
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

render(html`<root-app />`, document.body);

setTimeout(() => { 
  handlers.go('/')
}, 300)