const template =document.createElement("template")
template.innerHTML = `
    <div class="container">
    <h1 >
        NOT FOUND
    </h1>
    </div>
`


export default function renderPage(){
    const page = (template.content.cloneNode(true))
    return page
}