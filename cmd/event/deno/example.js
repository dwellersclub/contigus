
/*
    This script once installed will have an id attached to it
    Once updated all listener will be removed 

    default export methods to set up a script.
    a script is configured by default to listen to one or multiple projects.
    Some events within the projects are standards.
    Some others will be specific to each project.
*/
export const setUp = ( { projects, events, env , jobs}) => {

    events.on("app.code.**", (event) => { 
        console.log("my event", event)
    })

}