import {h, Component} from 'preact'
import { store } from '../tools'

export class StoreSubscriber extends Component {

  constructor(props){
    super(props)
    this.mountStore = this.mountStore.bind(this)
    this.unmountStore = this.unmountStore.bind(this)
    this.state = {};
    this.mountStore(props)
  }

  componentWillUnmount(){this.unmountStore()}

  mountStore(props){
    let stateMapping = this.getStateMapping ? this.getStateMapping() : []
    
    const {values, unmount, setUpdater} = store.mount(this, stateMapping)
    setUpdater((values) => {
      this.setState(Object.assign(this.state, values))
    })

    this.state = Object.assign(this.state, values)
    this.unmount = unmount
    console.log('connected',this.unmount)
  }

  unmountStore(){
    if(this.unmount){ this.unmount()}
    console.log('disconnected',this)
  }

}