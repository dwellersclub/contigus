import {h} from 'preact'

import { Calendar } from './screen/calendar'
import { PropertyEditor } from './screen/edit'
import { ImageEditor } from './screen/image'
import { RateEditor } from './screen/rates'
import { Search } from './screen/search'
import { Dashboard } from './screen/dashboard'

export default function({store, router}) {
    
    store.addState('inventory', {})

    const currentRoutes = {
        '/inventory/dashboard' : () =>  <Dashboard  store={store} />,
        '/inventory/search' : () =>  <Search store={store} />,
        '/inventory/property/:id/view' : ({ params }) =>  <PropertyEditor id={params.id} store={store} />,
        '/inventory/property/:id/calendar' : ({ params }) =>  <Calendar  id={params.id} store={store} />,
        '/inventory/property/:id/image' : ({ params }) =>  <ImageEditor  id={params.id} store={store} />,
        '/inventory/property/:id/rates' : ({ params }) =>  <RateEditor  id={params.id} store={store} />,
    }

    router.addRoutes(currentRoutes)

    return <Dashboard store={store} />
}
