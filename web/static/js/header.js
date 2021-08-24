import { getAuthUser } from "./auth.js";

const header = document.querySelector("header");
const authUser = getAuthUser();
const authenticated = authUser !== null;

header.innerHTML = `
    <nav>
        <a href="/">HOME</a>
        ${
          authenticated
            ? `<a href="/users/${authUser.username}">profile</a>
        <button id="logout-button">Logout</button>
        `
            : ""
        }
        </nav>
`;
if (authenticated) {
  const logoutButton = header.querySelector("#logout-button");
  logoutButton.addEventListener("click", onLogoutButtonClick);
}

function onLogoutButtonClick(ev) {
  const button = ev.currentTarget;
  button.disabled = true;
  localStorage.clear();
  location.reload();
}
