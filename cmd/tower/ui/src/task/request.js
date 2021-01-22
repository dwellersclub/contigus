const defaultOptions = {
    headers: { 'Content-Type': 'application/json;charset=UTF-8' },
    mode: 'cors',
}

export class HttpRequestService {
    constructor(config) {
        this.request = this.request.bind(this)
        this.config = { ...config }
    }

    request(url, method, body, options= {}) {
        if (method === 'FPOST') {
            method = 'POST'
            body = new URLSearchParams(body)
        }

        return fetch(`${url}`, { ...this.defaultOptions, ...options, method, body })
            .then((resp) => {
                const contentType = resp.headers.get('Content-Type')

                if (resp.status === 401) {
                    if (url.includes('userinfo')) {
                        return { username: 'visitor' }
                    }
                    return { errMsg: 'error_logout' }
                }

                if (!contentType || !contentType.includes('application/json')) {
                    return resp.status < 300
                        ? {}
                        : { errMsg: `error_request_${resp.status}` }
                }

                const data = {}

                try {
                    return resp.json().then((result) => {
                        if (Array.isArray(result)) {
                            return result
                        }
                        return { ...data, ...result }
                    })
                } catch (ex) {
                    console.error(ex)
                    return { errMsg: ex.message }
                }
            })
            .catch((err) => {
                console.error(err)
                return { errMsg: err.message }
            })
    }
}
