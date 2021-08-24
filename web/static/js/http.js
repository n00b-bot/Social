import {isAuthUser} from './auth.js'
import {parseJSON, stringifyJSON} from './lib/json.js';
import {isObject} from './utils.js';
export function doGet(url,headers) {
    console.log(localStorage.getItem('token'))
    return fetch(url,{
        headers: Object.assign(defaultHeaders(),headers)
    }).then(parseResponse)
}


export function doPost(url,body,headers){
    const inti = {
        method: 'POST',
        headers: defaultHeaders(),
    }
    if (isObject(body)){
        inti['body'] = stringifyJSON(body)
        inti.headers['content-type']= "application/json; charset=utf-8"
    }
    
    Object.assign(inti.headers,headers)
    return fetch(url,inti).then(parseResponse)
}

function defaultHeaders() {
    return isAuthUser() ? {'Authorization': 'Bearer '+localStorage.getItem('token')} : {};
}


/**
 * 
 * @param {Response} response 
 */

async function parseResponse(response){ 
    if (response.status ==204){
        return
    }
    const test = /** @type {Promise} */ (response).text()
    const  body = parseJSON(await test)
    if (!response.ok){  
        const msg = String(body.error)
        const err = new Error(msg)
        err.name= msg.toLowerCase().split(' ').map(word => {
                return word.charAt(0).toUpperCase() + word.slice(1)
        }).join('')+"Error"
        err['statusCode'] = response.status
        err['statusText'] = response.statusText
        err['url'] = response.url
        throw err
    }
    console.log(body)
    return body
}

export function subscribe(url,cb) {
    if (isAuthUser()) {
        const _url = new URL(url,location.origin)
        _url.searchParams.set('token',localStorage.getItem('token'))
        url=_url.toString()
    }
    const eventSource = new EventSource(url)
    eventSource.onmessage = ev => {
        try {
            cb(parseJSON(ev.data))

        }catch (_) {}
    }
    return ()=>{
        eventSource.close()
    }
}