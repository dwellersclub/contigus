import approve from 'approvejs'

export const validateForm = (form, rules,) => {
  const errors = {}
  const values = {}

  for (let index = 0; index < form.elements.length; index++) {
    const element = form.elements[index]
    const rule = rules[element.name]
    if (rule) {
      values[element.name] = element.value
      const validationResult = approve.value(element.value, rule)
      if (!validationResult.approved) {
        errors[element.name] = `msg_${form.id}_form_${element.name}_required`
      }
    }
  }

  return { errors, values }
}

export const validateFormElement = (element, rules,) => {
  const errors = {}
  const values = {}
  const rule = rules[element.name]
  values[element.name] = element.value
  if (rule) {
    const validationResult = approve.value(element.value, rule)
    if (!validationResult.approved) {
      errors[element.name] = `msg_${element.form.id}_form_${element.name}_required`
    }
  }

  return { errors, values }
}