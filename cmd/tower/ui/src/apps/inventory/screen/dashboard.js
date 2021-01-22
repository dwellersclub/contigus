import { h, Fragment } from 'preact'
import { useState, useEffect } from 'preact/hooks'
import PropTypes from 'prop-types'

export const Dashboard = ({ store }) => {
    const { values, unmount, setUpdater } = store.mount(Dashboard, [
        'inventory:/inventory',
    ])
    const [state, setState] = useState(values)

    setUpdater(setState)

    useEffect(() => unmount, [])

    return (
        <Fragment>
        <div className="clr-row">
        <div className="clr-col-1"></div>
        <div className="clr-col-10">
            <cds-form-group layout="vertical" cds-layout="container:xl">
            <p cds-text="heading">Find properties</p>
            <br />
            <div cds-layout="grid cols@lg:3 gap:lg" >
                <cds-search >
                  <label>Property</label>
                  <input type="search" onInput={console.log}/>
                </cds-search>
                <cds-button onClick={console.log}>Search</cds-button>
            </div>
            </cds-form-group>
        </div>
        <div className="clr-col-1"></div>
    </div>
    <br />
        <div  className="clr-row">
            <div className="clr-col-1"></div>
            <div className="clr-col-10">
                <table className="table">
                    <thead>
                        <tr>
                            <th>Decimal</th>
                            <th>Hexadecimal</th>
                            <th>Binary</th>
                            <th>Roman Numeral</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td>1</td>
                            <td>1</td>
                            <td>1</td>
                            <td>I</td>
                        </tr>
                    </tbody>
                </table>
            </div>
            <div className="clr-col-1"></div>
        </div>
    </Fragment>
    )
}

Dashboard.stateTypes = {
    $go: PropTypes.func,
}
