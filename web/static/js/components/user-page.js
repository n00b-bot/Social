import {doGet,doPost} from '../http.js';
import renderAvatarHTML from '../components/avatar.js'
import { isAuthUser } from '../auth.js';
import renderPost from './post.js'
const PAGE_SIZE =3
const template = document.createElement("template");
template.innerHTML = `
    <div class="user-wrapper">
        <div class="container wide">
            <div id="user-div"></div>
        </div>
    </div>
    <div class="container">
        <h2>POSTS</h2>
        <ol id="posts-div"></ol>
        <button id="load-more-button" class="load-more-post-button">LOAD MORE</button>
    </div>
`

export default async function renderPage(params) {
    const [user,posts] = await Promise.all([http.fetchUser(params.username),http.fetchPosts(params.username)])
    const page = template.content.cloneNode(true)
    const userDiv = page.getElementById("user-div")
    const postsDiv = page.getElementById("posts-div")
    const loadMoreButton = /** @type {HTMLButtonElement} */page.getElementById("load-more-button")
    if (posts!==null){
    for (const post of posts) {
        post.user =user
        postsDiv.appendChild(renderPost(post))
    }}
    const onLoadMoreButtonClick = async () => {
    const lastPost = posts[posts.length - 1];
    console.log(lastPost.id)
    const newPost = await http.fetchPosts(params.username,lastPost.id)
    posts.push(...newPost);
    for (const post of newPost) {
      post.user =user
      postsDiv.appendChild(renderPost(post));
    }
  };
   loadMoreButton.addEventListener("click", onLoadMoreButtonClick);
    

    userDiv.appendChild(renderUser(user))
    return page
}
/**
 * @param {import('../type.js').UserProfile} username
*/
export function renderUser(user) {
        const authenticated = isAuthUser()
        const div = document.createElement("div")
        
        div.className='user-profile'
        div.innerHTML=`
            <a href='/users/${user.username}'>
           ${renderAvatarHTML(user)} 
           </a>
           <div class="center">
                <div>
                
                <div>
                <h1>${user.username}</h1>
                ${user.Followeed ? `<span class="badge">Follows you</span>` :''}
                </div>
                </div>
                <div class="user-stats">
                    <a href="/users/${user.username}/followers">
                    ${user.FolloweesCount}
                    followers</a>
                    <a href="/users/${user.username}/followees"><span class="followers_count-span">${user.FollowersCount} </span> followees</a>
                </div>
                </div>
        
                ${authenticated && !user.Me ? `
                <button class="follow-button"> 
                    ${user.Following ? 'Following':'Follow'}

                </button>`:
                ''}
            `
        const followButton = /** @type {HTMLButtonElement} */ (div.querySelector('.follow-button'))
        const followersCountSpan = div.querySelector('.followers_count-span')
        if (followButton!==null){
            const onFollowButtonClick = async ()=>{
                    followButton.disabled =true
                    try {
                        const out = await http.toggleFollow(user.username)
                        followersCountSpan.innerHTML=String(out.FollowersCount)
                        followButton.textContent = out.Following ? 'Following':'Follow'
                    }
                    catch (e) {
                            console.log(e)
                    }finally {
                        followButton.disabled = false
                    }
            }
            followButton.addEventListener('click',onFollowButtonClick)
        }
        return div
}



const http = {
    /**
     * 
     * @param {string} username 
     * @returns {Promise<import('../type.js').UserProfile>}
     */
    fetchUser:username => doGet(`/api/users/${username}`),
    /**
     * @param {string} username 
     * @param {bigint} before
     * @returns {Promise<import('../type.js').Post[]>}
     */
    fetchPosts:(username,before) => {
        before=typeof before !== 'undefined'?before:""
       return doGet(`/api/users/${username}/posts?before=${before}&last=${PAGE_SIZE}`)
    },
    toggleFollow:(username) => doPost(`/api/users/${username}/toggle_follow`)
}