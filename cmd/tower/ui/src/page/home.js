import {h, Component} from 'preact'
import { useState} from 'preact/hooks';
import { store } from '../tools'

export default (router) => {

  const getInventoryApp =(path) => {
    let Node = null
    const App = () => {
      const [loaded, setLoaded] = useState(false)
      
      import(`../apps/inventory/inventory.js`).then((M) => {
        setLoaded(true)
        Node = M.default
      })

      return !loaded ? <div>Loading</div>: <Node store={store} router={router}/>
    }
    return App
  }

  const currentRoutes = {
    'p|/' : () =>  <HomePage />,
    'p|/inventory' : () => {
      const Inventory = getInventoryApp()
      return  <Inventory />
    } ,
  }

  router.addRoutes(currentRoutes)
}

class HomePage extends Component {
  render() {
    return <div>My Home Page Test</div>;
  }
}