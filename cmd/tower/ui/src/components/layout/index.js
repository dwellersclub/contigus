import { h } from 'preact'
import { store } from '../../tools'
import { useState, useEffect } from 'preact/hooks'
import { StoreSubscriber } from '../../tools/components'

import PropTypes from 'prop-types'

import '@clr/core/forms/register.js'
import '@clr/core/button/register.js'
import '@clr/core/alert/register.js'
import '@clr/core/select/register.js'
import '@clr/core/checkbox/register.js'
import '@clr/core/input/register.js'
import '@clr/core/date/register.js'
import '@clr/core/tag/register.js'
import '@clr/core/search/register.js'
import '@clr/core/datalist/register.js'

import {
    ClarityIcons,
    userIcon,
    nodeGroupIcon,
    headphonesIcon,
    cogIcon,
} from '@clr/core/icon'

ClarityIcons.addIcons(userIcon, nodeGroupIcon, headphonesIcon, cogIcon)

export class LayoutConfig extends StoreSubscriber {
    getStateMapping() {
        return ['page:/layout/page']
    }

    render({}, { page }) {
        const { title, description } = page

        document.title = title

        var metaDesc = document.querySelector('meta[name="description"]')
        if (metaDesc) {
            metaDesc.setAttribute('content', description)
        }

        var metaTitle = document.querySelector('meta[name="title"]')
        if (metaTitle) {
            metaTitle.setAttribute('content', title)
        }

        return null
    }
}

export class Layout extends StoreSubscriber {
    getStateMapping() {
        return ['page:/layout/page', 'header:/layout/header']
    }

    render({}, { header }) {
        return (
            <div className="main-container">
                {header.enable && (
                    <header className="header header-7">
                        <Logo />
                        <HeaderNav />
                        <HeaderSearch />
                        <div className="header-actions">
                            <UserSettings />
                        </div>
                    </header>
                )}
                <SubNav />
                <div className="content-container">
                    <div className="content-area">
                        <Content />
                    </div>
                </div>
            </div>
        )
    }
}

export class Content extends StoreSubscriber {
    getStateMapping() {
        return ['childNode:/ui/childNode']
    }
    render() {
        return window.ChildNode
    }
}

const Logo = () => {
    const { values, unmount, setUpdater } = store.mount(Logo, [':/layout/logo'])
    const [state, setState] = useState(values)

    setUpdater(setState)

    useEffect(() => unmount, [])

    const { $go } = state

    return (
        state.enable && (
            <div className="branding">
                <a className="nav-link" onClick={() => $go('/')}>
                    <span className="title">EdgeRetreats Admin</span>
                </a>
            </div>
        )
    )
}

Logo.stateTypes = {
    $go: PropTypes.func,
}

const HeaderNav = () => {
    const { values, unmount, setUpdater } = store.mount(HeaderNav, [
        ':/layout/nav',
    ])
    const [state, setState] = useState({ ...values })

    setUpdater(setState)

    useEffect(() => unmount, [])

    const { items, selected, $changePage } = state

    return (
        state.enable && (
            <div className="header-nav">
                {items.map((item) => {
                    const onClick = () => {
                        $changePage(item)
                    }
                    return (
                        <a
                            className={`nav-link nav-text ${
                                selected === item.id ? 'active' : ''
                            }`}
                            onClick={onClick}
                        >
                            {item.label}
                        </a>
                    )
                })}
            </div>
        )
    )
}

HeaderNav.stateTypes = {
    $changePage: PropTypes.func,
}

const HeaderSearch = () => {
    const { values, unmount, setUpdater } = store.mount(HeaderNav, [
        ':/layout/search',
    ])
    const [state, setState] = useState({ ...values })

    setUpdater(setState)

    useEffect(() => unmount, [])

    return (
        state.enable && (
            <form className="search">
                <label for="search_input">
                    <input type="text" placeholder="Search for keywords..." />
                </label>
            </form>
        )
    )
}

const Settings = () => (
    <a href="javascript://" className="nav-link nav-icon" aria-label="settings">
        <cds-icon shape="cog" />
    </a>
)

const UserSettings = () => (
    <div className="dropdown">
        <a className="nav-text dropdown-toggle" aria-label="open user profile">
            john.doe@vmware.com
        </a>
        <div className="dropdown-menu">
            <a href="javascript://" className="dropdown-item">
                Preferences
            </a>
            <a href="javascript://" className="dropdown-item">
                Log out
            </a>
        </div>
    </div>
)

const SubNav = () => {
    const { values, unmount, setUpdater } = store.mount(Logo, [
        ':/layout/subnav',
    ])
    const [state, setState] = useState(values)

    setUpdater(setState)

    useEffect(() => unmount, [])

    return state.enable && <nav class="subnav" />
}

const Alert = () => (
    <div className="alert alert-danger" role="alert">
        <div className="alert-items">
            <div className="alert-item static">
                <div className="alert-icon-wrapper">
                    <cds-icon
                        className="alert-icon"
                        shape="exclamation-circle"
                    ></cds-icon>
                </div>
                <span className="alert-text">
                    This alert is at the top of the page.
                </span>
            </div>
        </div>
    </div>
)

const PageHeader = () => {}
const Footer = () => {}
