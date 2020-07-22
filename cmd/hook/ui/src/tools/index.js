import { setCookie, getCookie } from './web'
import { getLanguage, loadLanguages, setLanguage, text, escapeHtml, capitalize, formatDate, formatDateRel, formatDateTime } from './i18n'
export { buildActions } from './actions'
export { store } from './store'
export { createRoutes } from './routes'
export { combineReducers, getMethods } from './utils'

export const web = { setCookie, getCookie }
export const i18n = { getLanguage,loadLanguages, setLanguage,text,escapeHtml, capitalize, formatDate,formatDateRel,formatDateTime}