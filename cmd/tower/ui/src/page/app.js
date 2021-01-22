import {h, Fragment, Component} from 'preact'
import { LayoutConfig, Layout } from '../components/layout'
import '../scss/contigus.scss'

export class Application extends Component  {
  render() {
    return <Fragment>
      <LayoutConfig />
      <Layout />
    </Fragment>;
  }
}