export const getDefaultState = (defaultLanguage) => {
  return {
    user: {name: 'anonymous'},
    location: {currentPath: ''},
    ui: { childNode: null, language: defaultLanguage, errMsg: null},
    layout: {
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
        debug : false,
        contentFull:false,
        layoutFluid:false,
        containerCentered:true,
        layoutHideTopbar: false,
        layoutTopbarTransparent: false
      },
      header: {
        enable: true,
      },
      logo: {
        enable: true,
      },
      nav: {
        items: [
          {id: 'home', label: 'Home', desc: '', path: '/'},
          {id: 'sale', label: 'Sale', desc: '', path: '/sale'},
          {id: 'inventory', label: 'Inventory', desc: '', path: '/inventory'},
          {id: 'reporting', label: 'Reporting', desc: '', path: '/reporting'},
          {id: 'system', label: 'System', desc: '', path: '/system'}
        ],
        selected: 'home',
        enable: true,
      },
      search: {
        enable: true,
      },
      sidebar: {
        enabled: false,
      },
      subnav: {
        enabled: false,
      }
    },
    log: {debug: false}
  }
}