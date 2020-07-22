let dict = {}

const defaultLanguage = 'en-GB'
let currentLanguage = 'en-GB'

export const loadLanguages = (languages) => {
  dict = Object.keys(languages).reduce((result, key) => {
    const list = languages[key]
    let base = {}
    Object.keys(list).map((i) => {
      if (i === key) {
        result[i] = list[i]
        base = list[i]
      } else {
        result[`${key}-${i}`] = Object.assign({}, base, list[i])
      }
    })
    return result
  } , {})
}

export const text = (key, args) => {
  let content = null
  const items = [currentLanguage]

  if (currentLanguage !== defaultLanguage) {
    items.push(defaultLanguage)
  }

  items.forEach(
    (language) => {
      const selectedDict = dict[language]
      if (!content && selectedDict && selectedDict.hasOwnProperty(key)) {
        content =  selectedDict[key](args)
      }
    }
  )
  return content ? content : key
}

export const setLanguage = (language) => {currentLanguage = language}
export const getLanguage = () => currentLanguage

export const escapeHtml = (str) => {
  const div = document.createElement('div')
  div.appendChild(document.createTextNode(str))
  return div.innerHTML
}

const dateFormatOptions = { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' }

export  const capitalize = (string) => (string || string.length > 0) ? string[0].toUpperCase() + string.slice(1) : string

export const formatDate = (currentDate) => currentDate.toLocaleDateString('en-US', dateFormatOptions)
export const formatDateTime = (currentDate) => currentDate.toLocaleTimeString('en-US', dateFormatOptions)

const minute = 60,
  hour = minute * 60,
  day = hour * 24,
  week = day * 7,
  month = week * 4

export const formatDateRel = (target, text, locale) => {
  const delta = (new Date()).getTime() - target.getTime()
  let msg = ''
  if (delta < 30) {
    msg = text('fmt_date_just_now')
  } else if (delta < minute) {
    msg = text('fmt_date_secong_ago' , {DELTA: delta})
  } else if (delta < 2 * minute) {
    msg = text('fmt_date_one_minute_ago')
  } else if (delta < hour) {
    msg = text('fmt_date_minutes_ago' , {DELTA: Math.floor(delta / minute)})
  } else if (Math.floor(delta / hour) == 1) {
    msg = text('fmt_date_one_hour_ago')
  } else if (delta < day) {
    msg = text('fmt_date_hours_ago' , {DELTA: Math.floor(delta / hour)})
  } else if (delta < day * 2) {
    msg = text('fmt_date_yesterday')
  } else if (delta < month) {
    msg = text('fmt_date_weeks_ago' , {DELTA: Math.floor(delta / week)})
  } else {
    msg = target.toLocaleDateString(locale, dateFormatOptions)
  }
  return msg
}
