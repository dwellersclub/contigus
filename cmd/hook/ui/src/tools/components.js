import { LitElement } from 'lit-element';
import { store } from '../tools'

export class BaseLitElement extends LitElement {
  static get properties() { return {
    state: { type: Object, reflect: false }
  };}

  constructor(){
    super()
    this.mountStore = this.mountStore.bind(this)
    this.unmountStore = this.unmountStore.bind(this)
    this.state = { };
  }

  connectedCallback(){
    super.connectedCallback()
    this.mountStore()
  }

  mountStore(){
    let stateMapping =  this.getStateMapping ? this.getStateMapping() : []
    this.closer = store.mount(this, (values) => {
      this.state = Object.assign({}, this.state, values)
    }, stateMapping)()
    console.log('connected',this)
  }

  unmountStore(){
    if(this.closer){
      this.closer()
    }
    console.log('disconnected',this)
  }

  disconnectedCallback(){
    super.disconnectedCallback()
    this.unmountStore()
  }

}