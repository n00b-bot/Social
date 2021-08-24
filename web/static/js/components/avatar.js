/**
 * @param {import('../type.js').User} user
*/
export default function renderAvatarHTML(user){
    console.log(typeof user.avatar_url !== "undefined");
    return user.avatar_url !==null && typeof user.avatar_url !== "undefined"
    ? `<img class="avatar" src="${user.avatar_url}" alt="${user.username}'s avatar">`
    :`<span class="avatar" data-initial="${user.username[0]}"></span>`;
}