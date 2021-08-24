import { createRouter } from "./lib/router.js";
import { guard } from "./auth.js";

let currentPage
const disconnectEvent = new CustomEvent('disconnect')
const r = createRouter();
r.route("/", guard(view("home"), view("login")));
r.route(/\//, view("not-found"));
r.subscribe(renderInfo(document.querySelector("main")));
r.install();
function renderInfo(target) {
  return async (result) => {
    if (currentPage instanceof Node){
      currentPage.dispatchEvent(disconnectEvent);
      target.innerHTML=''

    }
    target.innerHTML = "";
    currentPage=await result
    target.appendChild(currentPage);
  };
}

function view(name) {
  return (...args) =>
    import(`./components/${name}-page.js`).then((m) => m.default(...args));
}
