import {h, Component} from 'preact'

export default (routes) => {
  const currentRoutes = {
    'p|/' : () =>  <HomePage />,
  }
  Object.assign(routes, currentRoutes)
}

class HomePage extends Component {
  render() {
    return <div>Home Page</div>;
  }
}