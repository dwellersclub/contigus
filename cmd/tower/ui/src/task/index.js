
import { getMethods, combineReducers } from '../tools'
import { LocalService,  LocalServiceReducer , LayoutReducer, LayoutService} from './service'

const allServices = [new LocalService(), new LayoutService()]

const methods = allServices.reduce((items, currentService) => {
  Object.assign( items, getMethods(currentService))
  return items
}, {})

export const actions = {
  names() { return Object.keys(methods) },
  action(key) {
    const method = methods[key]
    if (!method) {
      console.error(`method  ${key} not found`)
      return null
    }
    return method
  }
}

export const actionReducers = combineReducers([new LocalServiceReducer(), new LayoutReducer()])
