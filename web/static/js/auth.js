import {parseJSON} from './lib/json.js'

export  function getAuthUser(){
    const authUserRaw = localStorage.getItem('auth_user')
    if (authUserRaw === null) {
        return null
    }
    if (localStorage.getItem('token')===null){
        return null
    }
    const expireAtRaw = localStorage.getItem('expire_at')
    if (expireAtRaw === null){
        return null
    }
    const expireAt = new Date(expireAtRaw)
    if (isNaN(expireAt.valueOf())|| expireAt <= new Date()){
        return null
    }
    try {
        return parseJSON(authUserRaw)
    }
    catch (_){
        return null
    }
}

export function isAuthUser(){
    return getAuthUser() !== null 
}

export  function guard(fn1, fn2){
    return (...args) =>  isAuthUser() ? fn1(...args) : fn2(...args)
}