import { doGet } from "../http.js";
import { renderUser } from "./user-page.js";
const template = document.createElement("template");
template.innerHTML = `
    <div class="container">
        <h1><span id ='username-outlet'></span>'s followers </h1>
        <div id="followers-outlet" class="followers-wrapper users-wrapper">
        <button id="search-more-button">Load More</button>
        </div>
    </div>
`;
const PAGE_SIZE = 3;

export default async function renderFollowersPage(params) {
  const page = template.content.cloneNode(true);
  const users = await fetchFollowers(params.username)
  const followersOutlet = page.getElementById("followers-outlet");
  const searchMoreButton = page.getElementById("search-more-button")
const usernameOutlet = page.getElementById("username-outlet");
usernameOutlet.textContent = params.username;
  for (const user of users) {
    followersOutlet.appendChild(renderUser(user));
  }
  const onLoadMoreButtonClick = async () => {
    const lastUser = users[users.length - 1];
    console.log(lastUser.id);
    const newUser = await fetchUsers(searchQuery,lastUser.username);
    users.push(...newUser);
    for (const user of newUser) {
      followersOutlet.appendChild(renderFollowers(user));
    }
  };
  searchMoreButton.addEventListener('click', onLoadMoreButtonClick)
  return page;
}

function fetchFollowers(username, after = "") {
  return doGet(`/api/users/${username}/followers?search=&after=${after}&first=${PAGE_SIZE}`);
}
