import {h, Component} from 'preact'

export default (routes) => {
  const currentRoutes = {
    'p|/signup' : () => <LoginPage />,
  }
  Object.assign(routes, currentRoutes)
}

class LoginPage extends Component {
  render() {
    return <div>Login Page</div>;
  }
}