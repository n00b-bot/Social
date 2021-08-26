import { doGet, doPost, subscribe } from "../http.js";
import renderPost from "./post.js";
import renderAvatarHTML from "./avatar.js";
import { escapeHTML } from "../utils.js";

import { getAuthUser } from "../auth.js";
const PAGE_SIZE = 3;
const template = document.createElement("template");
template.innerHTML = `
    <div class="post-wrapper">
        <div class='container'>
            <div id="post-outlet"></div>
        </div>
    </div>
    <div class="container">
        <ol id="comments-outlet"></ol>
        <form id="comment-form" class="comment-form">
            <textarea placeholder="Say something ..." maxlength="480" required></textarea>
             <button id="comment-button">Comment</button>
             
        </form>
        <button id="load-more-button">Load More</button>
    </div>
`;
export default async function renderPostPage(params) {
  const postID = BigInt(params.postID);
  const [post, comments] = await Promise.all([
    fetchPost(postID),
    fetchComment(postID, ""),
  ]);
  const page = template.content.cloneNode(true);
  const postOutlet = page.getElementById("post-outlet");
  const commentsOutlet = page.getElementById("comments-outlet");
  postOutlet.appendChild(renderPost(post));
  const commentForm = page.getElementById("comment-form");
  const commentFormText = commentForm.querySelector("textarea");
  const commentFormButton = page.getElementById("comment-button");
  const loadMoreButton = page.getElementById("load-more-button");

  const onLoadMoreButtonClick = async () => {
    try {
      const lastComment = comments[comments.length - 1];
      const newCommentItem = await fetchComment(postID, lastComment.ID);
      comments.push(...newCommentItem);
      for (const comment of newCommentItem) {
        commentsOutlet.appendChild(renderComment(comment));
      }
    } catch (e) {
      console.log(e);
    }
  };

  const onCommentFormSubmit = async (ev) => {
    ev.preventDefault();
    const content = commentFormText.value;
    commentFormButton.disabled = true;
    commentFormText.disabled = true;
    try {
      console.log(content);
      const comment = await createComment(post.id, { content });
      comments.unshift(comment);
      commentsOutlet.insertAdjacentElement("afterend", renderComment(comment));
    } catch (e) {
      console.log(e);
    } finally {
      commentFormButton.disabled = false;
      commentFormText.disabled = false;
    }
  };

  const onCommentArrive = (comment) => {
    comments.unshift(comment);
    commentsOutlet.insertAdjacentElement("afterend", renderComment(comment));
  };
  const unsubscribe = subscribeComment(post.id, onCommentArrive);

  const onPageDisconnect = () => {
    unsubscribe();
  };
  for (const comment of comments) {
    commentsOutlet.appendChild(renderComment(comment));
  }
  loadMoreButton.addEventListener("click", onLoadMoreButtonClick);
  commentFormButton.addEventListener("click", onCommentFormSubmit);
  page.addEventListener("disconnect", onPageDisconnect);
  return page;
}

export function renderComment(comment) {
  const { User } = comment;
  const ago = new Date(comment.CreateAt).toLocaleString();
  const li = document.createElement("li");
  li.className = "comments-item";
  li.innerHTML = `
        <article class="comment">
        <div class="comment-header">
            <a href="/users/${User.username}">
                ${renderAvatarHTML(User)}
                <span>${User.username}</span>
            </a>          
        </a>
        <time>${ago}</time>
        </div>
        <div class="post-content">${escapeHTML(comment.Content)}</div>
        <div class='post-controls'>
            <button class="${
              comment.Liked ? "like-button" : "unlike-button"
            }">
            <span>${comment.LikesCount}</span></button>
        
        </div>
        </article>
    
    `;
  const likeButton = li.querySelector("button");
  const likeCounts = likeButton.querySelector("span");
  console.log(likeCounts);
  console.log(likeButton);
  if (likeButton !== null) {
  const onLikeCommentClick = async () => {
      likeButton.disabled = true;
      try {
        const out = await toggleCommentLike(comment.ID);
        comment.LikesCount = out.LikesCount;
        comment.Liked = out.Liked;
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
    likeButton.addEventListener("click", onLikeCommentClick);
  }
 
  return li;
}

function fetchPost(postID) {
  return doGet("/api/posts/" + postID);
}

function fetchComment(postID, before) {
  return doGet(
    `/api/posts/${postID}/comments?before=${before}&last=${PAGE_SIZE}`
  );
}

async function createComment(postID, content) {
  const comment = await doPost(`/api/posts/${postID}/comments`, content);
  comment.User = getAuthUser();
  return comment;
}

function subscribeComment(postID, cb) {
  return subscribe(`/api/posts/${postID}/comments`, cb);
}

function toggleCommentLike(commentID) {
  return doPost(`/api/comments/${commentID}/toggle_like`);
}
