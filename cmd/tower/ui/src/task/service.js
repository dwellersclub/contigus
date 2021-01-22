export class LocalService {
    switchLanguage(languageCode) {
        return languageCode
    }
    locationChange(currentPath) {
        return currentPath
    }
    enableRemoteLogging(debug, debugToken) {
        return { debug, debugToken }
    }
    go(currentPath) {
        return currentPath
    }
    displayError(errorMsg) {
        return errorMsg
    }
    displayMsg(msg) {
        return msg
    }
    setUser(user) {
        return user
    }
}
export class LocalServiceReducer {
    setUser(user) {
        return (state) => {
            state.user = user
        }
    }

    switchLanguage(language) {
        return (state) => {
            state.ui = { ...state.ui, language }
        }
    }

    locationChange(lastLocation) {
        return (state) => {
            state.ui = { ...state.ui, lastLocation }
            setCookie(`ui_${this.appName}_location`, lastLocation, 1)
        }
    }

    enableRemoteLogging(details) {
        return (state) => {
            state.log = {
                ...state.log,
                debug: details.debug,
                debugToken: details.debugToken,
            }
        }
    }

    go(currentPath) {
        return (state) => {
            state.location = {
                ...state.location,
                currentPath,
                time: new Date().getTime(),
            }
        }
    }

    displayError(errorMsg, ctx) {
        return (state) => {
            state.ui.errMsg = errorMsg
            if (state.errMsgTimeout) {
                clearTimeout(state.errMsgTimeout)
            }
            if (errorMsg) {
                state.errMsgTimeout = setTimeout(
                    () => ctx.actions.displayError(null),
                    6000
                )
            }
        }
    }

    displayMsg(msg, ctx) {
        return (state) => {
            state.ui.msg = msg
            if (state.msgTimeout) {
                clearTimeout(state.msgTimeout)
            }
            if (msg) {
                state.msgTimeout = setTimeout(
                    () => ctx.actions.displayMsg(null),
                    2000
                )
            }
        }
    }
}

export class LayoutService {
    changePage(item) {
        return item
    }
}

export class LayoutReducer {
    changePage(item, ctx) {
        return (state) => {
            state.layout.nav.selected = item.id
            return () => {
                ctx.actions.go(item.path)
            }
        }
    }
}

export class PropertyService {
    constructor(requestService) {
        this.config = { ...config }
        this.request = requestService.request
    }

    searchProperties(request) {
        return this.request(url, 'POST', JSON.stringify(request))
    }
}

export class PropertyServiceReducer {
    changePage(item, ctx) {
        return (state) => {
            state.layout.nav.selected = item.id
            return () => {
                ctx.actions.go(item.path)
            }
        }
    }
}
