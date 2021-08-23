import {createRouter} from './lib/router.js'
import {guard} from './auth.js';

const r =createRouter()
r.route("/",guard(view('home'),view('login')))
r.route(/\//,view('not-found')
)
r.subscribe(renderInfo(document.querySelector("main")))
r.install()
function renderInfo(target){
    return async result=>{
        target.innerHTML = ''
        target.appendChild(await result)
       
    }
}

function view(name) {
    return (...args) => import(`./components/${name}-page.js`).then(m=>m.default(...args))
}