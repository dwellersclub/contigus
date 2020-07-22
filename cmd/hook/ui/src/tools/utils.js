const protoFuncToSkip = ['constructor', 'prototype', 'length', 'toString', 'apply',
  'call','caller','name', 'bind', 'arguments']

export function getMethods(proto) {
  if (!proto) {
    return []
  }
  const sources = [proto, Object.getPrototypeOf(proto)]
  const methods = {}

  sources.forEach((source) => {
    if (source) {
      const propertyNames = Object.getOwnPropertyNames(source)
      propertyNames.forEach((methodName) => {
        if (!protoFuncToSkip.includes(methodName)) {
          if (typeof source[methodName] === 'function') {
            if (!methods.hasOwnProperty(methodName)) {
              Object.assign(methods, { [methodName]:  source[methodName]})
            }
          }
        }
      })
    }
  })

  return methods
}

export function combineReducers(reducers) {
  const methods = {}
  reducers.forEach((reducer) => {
    Object.assign(methods, getMethods(reducer))
  })
  return methods
}
