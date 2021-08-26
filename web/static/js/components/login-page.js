import { doPost } from "../http.js";
import { stringifyJSON } from "../lib/json.js";

const template = document.createElement("template");
template.innerHTML = `
    <div class="container">
    <h1 >
        LOGIN PAGE
    </h1>
        <form id="login-form">
            <input type="text" class="form-control" placeholder="Email" autocomplete="email" value="john@dot.com" required>
            <button>Login</button>
        </form>
    </div>
`;

export default function renderPage() {
  const page = template.content.cloneNode(true);
  const loginForm = page.getElementById("login-form");
  loginForm.addEventListener("submit", onLoginFormSubmit, false);
  return page;
}

async function onLoginFormSubmit(ev) {
  ev.preventDefault();
  const from = ev.currentTarget;
  const input = from.querySelector("input");
  const button = from.querySelector("button");
  const email = input.value;
  input.disabled = true;
  button.disabled = true;
  try {
    const out = await http.login(email);
    localStorage.setItem("token", out.token);
    localStorage.setItem(
      "expire_at",
      typeof out.Expiration === "string"
        ? out.Expiration
        : out.Expiration.toJSON()
    );
    localStorage.setItem("auth_user", stringifyJSON(out.AuthUser));
    location.reload();
  } catch (err) {
    if (err.name === "UserNotFoundError") {
      if (confirm("User not found")) {
        Register(email);
      }
      return;
    }
    alert(err.message);
    setTimeout(input.focus);
  } finally {
    input.disabled = false;
    button.disabled = false;
  }
}

function saveLogin(data) {
  localStorage.setItem("token", data.token);
  localStorage.setItem("expire_at", data.Expiration);
  localStorage.setItem("auth_user", stringifyJSON(data.AuthUser));
}

async function Register(email, username) {
  username = prompt("Username", username);
  if (username === null) {
    return;
  }
  username = username.trim();
  try {
    await http.register(email, username);
 ;
    saveLogin(await http.login(email));
    location.reload();
  } catch (err) {
  ;
    alert(err.message);
  }
}

const http = {
  login: (email) => doPost("/api/login", { email }),
  register: (email, username) => doPost("/api/users", { email, username }),
};
