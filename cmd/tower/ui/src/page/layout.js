
import { LitElement, html, property, customElement } from 'lit-element';

@customElement('setup-layout')
export class SetupLayout extends LitElement {
  render() {
    return html`<div>
    <slot name="one"></slot>
    <slot name="two"></slot>
  </div>`;
  }
}