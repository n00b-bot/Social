import { doGet } from "../http.js";
import { renderUser } from "./user-page.js";
const template = document.createElement("template");
template.innerHTML = `
    <div class="container">
        <h1>Search</h1>
        <form id="search-form" class="search-form">
            <input type="search" name="q" placeholder="Search" autocomplete="off" autofocus="on"/>
            
        </form>
        <button id="search-more-button">Load More</button>
        <div id="search-result-outlet" class="search-result-wrapper users-wrapper">
        </div>
    </div>
`;
const PAGE_SIZE = 3;

export default async function renderSearchPage() {
  const url = new URL(location.toString());
  const searchQuery = url.searchParams.has("q")
    ? decodeURIComponent(url.searchParams.get("q")).trim()
    : "";
  const page = template.content.cloneNode(true);
  const users = await fetchUsers(searchQuery);
  const searchResultOutlet = page.getElementById("search-result-outlet");
  const searchMoreButton = page.getElementById("search-more-button")
  const searchForm = page.getElementById("search-form")
  const searchInput = searchForm.querySelector("input")
  const onSearchFormSubmit = ev => {
      if    (ev.keyCode ===  13){
      ev.preventDefault();
      const searchQuery =searchInput.value.trim();
      navigate("/search?q="+encodeURIComponent(searchQuery))}
  }

  setTimeout(() =>{
      searchInput.focus()
  })

  searchForm.addEventListener("keyup",onSearchFormSubmit)
  searchInput.value=searchQuery



  for (const user of users) {
    searchResultOutlet.appendChild(renderUser(user));
  }
  const onLoadMoreButtonClick = async () => {
    const lastUser = users[users.length - 1];
    console.log(lastUser.id);
    const newUser = await fetchUsers(searchQuery,lastUser.username);
    users.push(...newUser);
    for (const user of newUser) {
      searchResultOutlet.appendChild(renderUser(user));
    }
  };
  searchMoreButton.addEventListener('click', onLoadMoreButtonClick)
  return page;
}

function fetchUsers(search, after = "") {
  return doGet(`/api/users?search=${search}&after=${after}&first=${PAGE_SIZE}`);
}
