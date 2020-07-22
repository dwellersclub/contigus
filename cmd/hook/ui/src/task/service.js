
export class LocalService {
  switchLanguage(languageCode) { return languageCode }
  locationChange(currentPath) { return currentPath }
  enableRemoteLogging(debug, debugToken) { return {debug, debugToken}}
  go(currentPath) { return currentPath }
  displayError(errorMsg) { return errorMsg }
  displayMsg(msg) { return msg }
  setUser(user) { return user }
}

export class LocalServiceReducer {
  
  setUser(user) {
    return (state) => { Object.assign(state, { user })}
  }
  
  switchLanguage(language) {
    return (state) => {
      Object.assign(state.ui, { language })
    }
  }

  locationChange(lastLocation) {
    return (state) => {
      Object.assign(state.ui, { lastLocation })
      setCookie(`ui_${this.appName}_location`, lastLocation, 1)
    }
  }

  enableRemoteLogging(details) {
    return (state) => {
      Object.assign(state.log, { debug: details.debug, debugToken: details.debugToken })
    }
  }

  go(currentPath) {
    return (state) => {
      Object.assign(state.location, { currentPath, time: new Date().getTime() })
    }
  }

  displayError(errorMsg, ctx) {
    return (state) => {
      state.ui.errMsg = errorMsg
      if (state.errMsgTimeout) { clearTimeout(state.errMsgTimeout) }
      if (errorMsg) {
        state.errMsgTimeout = setTimeout(() => ctx.actions.displayError(null), 6000)
      }
    }
  }

  displayMsg(msg, ctx) {
    return (state) => {
      state.ui.msg = msg
      if (state.msgTimeout) { clearTimeout(state.msgTimeout) }
      if (msg) {
        state.msgTimeout = setTimeout(() => ctx.actions.displayMsg(null), 2000)
      }
    }
  }

}