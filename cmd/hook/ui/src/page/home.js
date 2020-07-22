import { LitElement, html, customElement } from 'lit-element';

export default (routes) => {
  const currentRoutes = {
    'p|/' : () => html`<home-page />`,
  }
  Object.assign(routes, currentRoutes)
}

@customElement('home-page')
class HomePage extends LitElement {
  render() {
    return html`Home Page`;
  }
}