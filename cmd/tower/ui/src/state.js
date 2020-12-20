export const getDefaultState = (defaultLanguage) => {
  return {
    user: {name: 'anonymous'},
    location: {currentPath: ''},
    ui: { childNode: null, language: defaultLanguage, errMsg: null},
    page : {
      title : "",
      rtl  : "",
      description : "",
      environment : "",
      preview : "",
      layoutDark : false,
      layoutClass : "",
      pageClass : "",
      bodyClass : "",
      debug : false
    },
    navbar: {
      prefix: 'navbar'
    },
    sidebar: {
      right: false,
      breakpoint: 'lg'
    },
    log: {debug: false}
  }
}