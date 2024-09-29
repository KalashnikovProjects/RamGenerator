let topRams;
let mapTopRams = {};
let loadingTopRams=true;

async function timeToTapHref() {
    if (!user) {
        document.querySelector("#time-to-tap").href = "/login"
        return;
    }

    document.querySelector("#time-to-tap").href = `/users/${user.username}`
}

async function indexRam(ram) {
    document.getElementById('ram').classList.add('target')
    let elem =  document.querySelector("#ram .popup-menu");
    let style =  `
            width: 1.5rem;
            height: 1.5rem;
            display: inline-block;
            `
    style += styleForAvatar(ram.user.avatar_url, ram.user.avatar_box)
    elem.innerHTML = `
             <h4 id="ram-description" class="ram-description">${ram.description}</h2>
             <div class="ram-user cursor-pointer" onclick="location.href='users/${ram.user.username}'">
                <div class="user-avatar" style="${style}"></div>
                <h7 id="ram-card-username" style="text-align: center">${ram.user.username}</h7>
             </div>
             <div id="ram-content" class="text-center ram-content ram-content-index">
                <img id="ram-clicker" class="ram-image" src="${ram.image_url}" alt="ram">
                <div id="taps-line" class="mt-3"><h3 id="ram-clicked">${ram.taps} тапов</h3></div>
             </div>
             <button id="close-button" style="right:1.5rem" class="up-button" onclick="closePopup()">
                 <svg xmlns="http://www.w3.org/2000/svg" fill="white" class="bi bi-x" viewBox="0 0 16 16">
                    <path d="M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708"/>
                </svg>
            </button>`
}

function closePopup() {
    document.querySelector("#ram .popup-menu").innerHTML = "";
    document.getElementById('ram').classList.remove('target')

    const url = new URL(location);
    if (url.searchParams.has("ram-id")) {
        url.searchParams.delete("ram-id");
        history.pushState({}, "", url);
    }
}

async function openRam(ram) {
    const url = new URL(location);
    url.searchParams.set("ram-id", `${ram.id}`);
    history.pushState({}, "", url);
    indexRam(ram)
}

async function checkHash() {
    if (loadingTopRams) {
        setTimeout(checkHash, 5);
        return;
    }
    const url = new URL(location);
    if (url.searchParams.get("ram-id")) {
        const url = new URL(location);
        let id = +url.searchParams.get("ram-id");
        if (!mapTopRams[id]) {
            closePopup()
            return
        }
        indexRam(mapTopRams[id])
    }
}

async function loadTopRams() {
    try {
        const response = await fetch(`${API_URL}/top-rams`, {
            mode: 'cors',
            method: 'GET',
        });
        if (response.ok) {
            topRams = await response.json();
            for (ram of topRams) {
                ram.user.avatar_url = ram.user.avatar_url || DEFAULT_AVATAR;
                mapTopRams[ram.id] = ram
            }
        } else {
            console.error(response.text())
            // TODO:
            // showError();
        }
    } catch (error) {
        console.log(error)
        // showError();
    }
}


loadTopRams().then(() => { if (topRams!==undefined) { loadingTopRams = false }});


function createRamElem(ram, place) {
    let style =  `
    width: 1.5rem;
    height: 1.5rem;
    display: inline-block;
    `
    style += styleForAvatar(ram.user.avatar_url, ram.user.avatar_box)
    return `<div class="col">
        <div class="ram-card ram-card-index ram-place-${place}" onclick="openRam(mapTopRams[${ram.id}])">
          <img src="${ram.image_url}" class="ram-card-image" alt="${ram.description}">
          <h3 style="text-align: center">${ram.taps} тапов</h3>
          <div class="ram-card-user cursor-pointer" onclick="location.href='users/${ram.user.username}';event.stopPropagation()">
            <div class="user-avatar" style="${style}"></div>
            <h7 id=="ram-card-username" style="text-align: center">${ram.user.username}</h7>
          </div>
        </div>
      </div>`
}

async function displayLoadingTopRams() {
    let res = "";
    for (let place=1; place<=5;place++) {
        res += `<div class="col"><div class="ram-card ram-card-index ram-place-${place}">
                <img src="/static/img/icon512.png" class="loading-image-top-rams loading-image rotating-image img-fluid wait-ram" alt="Загрузка...">
            </div></div>`
    }
    if (loadingTopRams) {
        document.getElementById("rams").innerHTML = `<div id="rams-list" class="row-gap-3 column-gap-3 rams-list row-cols-auto g-3 d-flex justify-content-center flex-wrap">${res}</div>`;
    }
    setTimeout(async () => {
        for (l of document.getElementsByClassName("loading-image-top-rams")) {
            l.classList.add("loading-image-visible")
            await sleep(300)
        }
    }, 10)

}

async function displayTopRams() {
    if (loadingTopRams) {
        setTimeout(displayTopRams, 5);
        return;
    }
    let res = "";
    let place = 1
    for (let ram of topRams) {
        res += createRamElem(ram, place);
        place++
    }
    document.getElementById("rams").innerHTML = `<div id="rams-list" class="row-gap-3 column-gap-3 rams-list row-cols-auto g-3 d-flex justify-content-center flex-wrap">${res}</div>`;
}