/**
 * @param {import('../type.js').Post} post
 */

import { doPost } from "../http.js";
import { escapeHTML } from "../utils.js";
import renderAvatarHTML from "./avatar.js";

export default function renderPost(post) {
  const { user } = post;
  const ago = new Date(post.create_at).toLocaleString();
  const li = document.createElement("li");
  console.log(typeof li);
  li.className = "post-item";
  li.innerHTML = `
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
            <button class="${   post.liked ? "like-button" : "unlike-button"}"
            title="${post.liked ? "Unlike" : "Like"}"
            aria-pressed="${post.liked}"
            aria-label="${post.likes_count} likes">
            <span class="likes-count">${post.likes_count}</span>
            </button>
            <a class="comment-link" href="/posts/${post.id}">${post.comments_count}</a>
        </div>
        </article>
    
    `;

  const likeButton = li.querySelector('button');

  const likeCounts= likeButton.querySelector('span');
  console.log(likeCounts);
  if (likeButton !== null) {
    const onLikeClick = async () => {
      likeButton.disabled = true;
      try {
        const out = await toggleLike(post.id);
        post.likes_count = out.LikesCount;
        post.liked =out.Liked;
        likeCounts.textContent = out.LikesCount;
        if (!out.Liked) {
          likeButton.className = "unlike-button";
        } else {
          likeButton.className = "like-button";
        }
        
      } catch (e) {
        console.log(e);
      } finally {
        likeButton.disabled = false;
      }
    };
    likeButton.addEventListener("click", onLikeClick);
  }
  return li;
}

function toggleLike(postID) {
  return doPost(`/api/posts/${postID}/toggle_like`);
}
