import {h, Component} from 'preact'
import { StoreSubscriber } from '../tools/components'

export class LayoutConfig extends StoreSubscriber {
  getStateMapping() { return ['page:/page']}

  render({} , {page}){

    const {title, description, layoutDark, 
      layoutClass, pageClass} = page

    document.title = title

    var metaDesc = document.querySelector('meta[name="description"]')
    if(metaDesc){ metaDesc.setAttribute('content', description)}

    var metaTitle = document.querySelector('meta[name="title"]')
    if(metaTitle){ metaTitle.setAttribute('content', title)}

    //set body class
    document.body.setAttribute("class", "")

    const classNames = []
    if(layoutDark){ classNames.push('theme-dark')}
    if(layoutClass){ classNames.push(layoutClass)}
    if(pageClass){ classNames.push(pageClass)}

    document.body.classList.add(...classNames)
    return null
  }

}

export class Layout extends Component {
  render({children}){
    return [children]
  }
}

const NavbarToggler = ({target}) => <button class="navbar-toggler" type="button" data-toggle="collapse" data-target={`#${target}`}>
  <span class="navbar-toggler-icon"></span>
</button>

const NavbarLogo = ({enabled, prefix, className, showTitle, dark, small}) =>  (enabled ? <a href="/" class={`${prefix}-brand {{ prefix }}-brand-autodark ${className}`}>
<img 
  src={`/static/logo${small ? '-small': ''}${dark ? '-white': ''}.svg`} 
  class={`navbar-brand-image ${showTitle ? 'mr-3': ''}`}></img>
{showTitle ? 'Dashboard': ''}
</a> : null)

const NavbarSide = () => { return null}
const NavbarMenu = () => { return null}

export class SideBar  extends StoreSubscriber {
  getStateMapping() { return ['sidebar:/sidebar']}
  render({enabled, right, breakpoint, className, background, dark, backgroundColor, hideBrand}){
    return enabled ? <aside class={`navbar navbar-vertical${right ? 'navbar-right': ''} navbar-expand-${breakpoint} ${dark ? 'navbar-dark' : 'navbar-light'} bg-${background} ${className}`}
       style={{ background: backgroundColor }} >
      <div class="container">
        <NavbarToggler target="navbar-menu"  />
        <NavbarLogo enabled={hideBrand} prefix="" className="" showTitle={true} dark={dark} small={false} />
        <NavbarSide className="d-lg-none" />
        <div class="collapse navbar-collapse" id="navbar-menu">
          <NavbarMenu />
          <div class={`d-none d-${breakpoint}-flex mt-auto`}>
            Lorem ipsum dolor sit amet, consectetur adipisicing elit. 
            Blanditiis, cupiditate doloribus dolorum impedit labore magni nisi nostrum rerum tempora! 
            Accusantium asperiores cum, est eum quas quia similique sunt ullam vel!
          </div>
        </div>
      </div>
    </aside>: null
  }
}