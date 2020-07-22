import { LitElement, html, property, customElement } from 'lit-element';
import './layout.js';

@customElement('setup-security')
export class SetupSecurity extends LitElement {
  render() {
    return html`<form>
        <input name="" value="" type="text">
        <input name="encryptionKey" value="" type="submit">
    </form>`;
  }
}