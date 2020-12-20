import {h, Fragment} from 'preact'
import { StoreSubscriber } from '../tools/components'
import { LayoutConfig, SideBar, Layout } from '../components/layout'
import '../scss/contigus.scss'
export class Application extends StoreSubscriber {
  getStateMapping() { return ['user:/user/*', 'childNode:/ui/childNode']}

  render() {
    const { user } = this.state
    return <Fragment>
      <LayoutConfig />
      <SideBar/>
      <Layout>{window.childNode}</Layout>
    </Fragment>;
  }
}