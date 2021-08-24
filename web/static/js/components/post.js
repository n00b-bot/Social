/**
 * @param {import('../type.js').Post} post
 */

import { escapeHTML } from '../utils.js'
import renderAvatarHTML from './avatar.js'

export default function renderPost(post){
    const {user} =post
    const ago =new Date(post.create_at).toLocaleString() 
    const li = document.createElement('li')
    li.className='post-item'
    li.innerHTML=`
        <article class="post">
        <div class="post-header">
            <a href="/users/${user.username}">
                ${renderAvatarHTML(user)}
                <span>${user.username}</span>
            </a>      
           
            <a href="/posts/${post.id}">
                
        </a>
        <time>${ago}</time>
        </div>
        <div class="post-content">${escapeHTML(post.content)}</div>
        <div class='post-controls'>
            <button class="like-button">${post.likes_count}</button>
            <a class="comment-link" href="/posts/${post.id}">${post.comments_count}</a>
        
        </div>
        </article>
    
    `
    return li
}