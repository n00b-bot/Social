import renderPost from "./post.js";
import { doGet, doPost, subscribe } from "../http.js";
import { getAuthUser } from "../auth.js";
const template = document.createElement("template");
template.innerHTML = `
    <div class="container">
    <h1>Timeline</h1>
    <form id="post-form" class="post-form">
        <textarea placeholder="Write Something" required maxlength="480"></textarea>
        <button class="post-form-button" >Publish</button>
    </form>
    <ol id="timeline-list" class="post-list">
    </ol>
    <button id="load-more-button" class="timeline-load-more-button">Load More</button>
    </div>
`;

const PAGE_SIZE = 3;

export default async function renderPage() {
  var timeline = await http.timeline("");
  const page = template.content.cloneNode(true);
  const postForm = page.getElementById("post-form");
  const postFormArea = postForm.querySelector("textarea");
  const postFormButton = postForm.querySelector("button");
  const timelineList = page.getElementById("timeline-list");
  const loadMoreButton = page.getElementById("load-more-button");

  /**
   * @param {Event} ev
   */
  const onPostFormSubmit = async (ev) => {
    ev.preventDefault();
    const content = postFormArea.value;
;
    postFormArea.disabled = true;
    postFormButton.disabled = true;
    try {
      const timelineItem = await http.publishPost({ content });
      timeline.unshift(timelineItem);
      timelineList.insertAdjacentElement(
        "afterbegin",
        renderPost(timelineItem.post)
      );
      postForm.reset();
      setTimeout(() => {
        postFormArea.focus();
      });
    } catch (err) {
  ;
      alert(err.message);
    } finally {
      postFormArea.disabled = false;
      postFormButton.disabled = false;
    }
  };
  const onLoadMoreButtonClick = async () => {
    const lastTimelineItem = timeline[timeline.length - 1];
    console.log(lastTimelineItem.id)
    const newTimelineItem = await http.timeline(lastTimelineItem.id);
    timeline.push(...newTimelineItem);
    for (const timelineItem of newTimelineItem) {
      timelineList.appendChild(renderPost(timelineItem.post));
    }
  };
  if (timeline !== null) {
    for (const timelineItem of timeline) {
      timelineList.appendChild(renderPost(timelineItem.post));
    }
  }else{
    timeline=[]
  }

  const onTimelineItemArrive = (timelineItem) => {
    timeline.unshift(timelineItem);
    timelineList.insertAdjacentElement(
      "afterbegin",
      renderPost(timelineItem.post)
    );
  };
  const unsubscribe = http.timelineSubscription(onTimelineItemArrive);
;
  postForm.addEventListener("submit", onPostFormSubmit);
  loadMoreButton.addEventListener("click", onLoadMoreButtonClick);
  page.addEventListener("disconnect", () => {
    unsubscribe();
  });
  return page;
}

const http = {
  publishPost: (input) =>
    doPost("/api/posts", input).then((timelineItem) => {
      timelineItem.post.user = getAuthUser();
      return timelineItem;
    }),
  timeline: (before) =>
    doGet(`/api/timeline?before=${before}&last=${PAGE_SIZE}`),
  timelineSubscription: (cb) => subscribe("/api/timeline", cb),
};
