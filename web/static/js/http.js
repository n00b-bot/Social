import {isAuthUser} from './auth.js'
import {parseJSON, stringifyJSON} from './lib/json.js';
import {isObject} from './utils.js';
export function doGet(url,headers) {
    return fetch(url,{
        headers: Object.assign(defaultHeaders,headers)
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
    console.log(fetch(url,inti).then(parseResponse))
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
    const body = JSON.parse(await response.text())
    if (!response.ok){
        const msg = string(body)
        const err = new Error(msg)
        err.name= msg.toLowerCase().split(' ').map(word => 
            (word.charAt(0).toUpperCase()+word.slice(1))).join('')+'Error'
        err['statusCode'] = response.status
        err['statusText'] = response.statusText
        err['url'] = response.url
        throw err
    }
    return body
}