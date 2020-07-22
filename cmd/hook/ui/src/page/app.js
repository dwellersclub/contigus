import { LitElement, html, property, customElement } from 'lit-element';
import { BaseLitElement } from '../tools/components'
import './layout.js';

@customElement('root-app')
export class Application extends BaseLitElement {
  getStateMapping() { return ['user:/user/*', 'childNode:/ui/childNode']}

  render() {
    const { user, childNode } = this.state
    return html`<setup-layout>
      <p slot="one">${user.name}</p>
      <p slot="two">${childNode}</p>
    </setup-layout>`;
  }
}