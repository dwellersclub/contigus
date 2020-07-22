import { LitElement, html, customElement } from 'lit-element';

export default (routes) => {
  const currentRoutes = {
    'p|/signup' : () => html`<login-page />`,
  }
  Object.assign(routes, currentRoutes)
}

@customElement('login-page')
class LoginPage extends LitElement {
  render() {
    return html`Login Page`;
  }
}