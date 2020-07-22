export const getDefaultState = (defaultLanguage) => {
  return {
    user: {name: 'anonymous'},
    location: {currentPath: ''},
    runtime: {},
    ui: { childNode: null, hidden: false, isCollapsed: false ,language: defaultLanguage, errMsg: null, progress : {value:100}},
    log: {debug: false}
  }
}