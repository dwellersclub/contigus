import {h, Component} from 'preact'
import { store } from '../tools'

export class StoreSubscriber extends Component {

  constructor(){
    super()
    this.mountStore = this.mountStore.bind(this)
    this.unmountStore = this.unmountStore.bind(this)
    this.state = {};
    this.mountStore()
  }

  componentWillUnmount(){this.unmountStore()}

  mountStore(){
    let stateMapping =  this.getStateMapping ? this.getStateMapping() : []
    this.closer = store.mount(this, (values, firstCall) => {
      if(firstCall) {
        Object.assign(this.state, values)
        return
      }
      this.setState(Object.assign(this.state, values))
    }, stateMapping)()
    console.log('connected',this.closer)
  }

  unmountStore(){
    if(this.closer){ this.closer()}
    console.log('disconnected',this)
  }

}