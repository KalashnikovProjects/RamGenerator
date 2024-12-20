"use strict";

let user;
const isTouchDevice =
    (('ontouchstart' in window) ||
        (navigator.maxTouchPoints > 0) ||
        (navigator.msMaxTouchPoints > 0));

function sleep(time) {
    return new Promise((resolve) => setTimeout(resolve, time));
}
let scrollTop = 0;

function hideScroll() {
    if (!isTouchDevice) {
        return
    }
    document.body.classList.add('no-scroll');
    scrollTop = window.scrollY;

    Object.assign(document.body.style, {
        position: 'fixed',
        width: `calc(100% - ${getScrollbarSize()}px)`,
        top: `${-scrollTop}px`
    });
}

function showScroll() {
    if (!isTouchDevice) {
        return
    }
    document.body.classList.remove('no-scroll');

    Object.assign(document.body.style, {
        position: '',
        width: '',
        top: ''
    });

    window.scrollTo(0, scrollTop);
}

function getScrollbarSize() {
    const outer = document.createElement('div');
    Object.assign(outer.style, {
        visibility: 'hidden',
        width: '100px',
        msOverflowStyle: 'scrollbar',
        overflow: 'scroll'
    });

    document.body.appendChild(outer);

    const inner = document.createElement('div');
    inner.style.width = '100%';
    outer.appendChild(inner);

    const scrollbarSize = outer.offsetWidth - inner.offsetWidth;
    outer.remove();

    return scrollbarSize;
}

async function loadUser() {
    try {
        const token = getCookie("token");
        if (!token) {
            user = null
            return user
        }

        user = sessionStorage.getItem("user");
        if (!!user) {
            user = JSON.parse(user)
            console.log(user)
            return user
        }

        const response = await fetch(`${API_URL}/me`, {
            mode: 'cors',
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Accept': 'application/json'
            },
        });

        if (response.ok) {
            user = await response.json();
            user.avatar_url = user.avatar_url || DEFAULT_AVATAR;
            sessionStorage.setItem("user", JSON.stringify(user));
            return user
        } else {
            const errorText = await response.text();
            console.error('Error response:', response.status, errorText);
        }
    } catch (error) {
        console.error('Fetch error:', error);
    }
}

function logOut() {
    sessionStorage.removeItem("user");
    deleteCookie("token");
    window.location.reload();
}

var loadingUser = true
loadUser().then(
    () => {
        loadingUser = false
    }
)

function styleForAvatar(avatar_url, avatar_box) {
    const [[x1, y1], [x2, y2]] = avatar_box;
    const size = Math.abs(y1 - y2);
    const moveSize = 1 - size;
    const posX = Math.min(x1, x2) / moveSize;
    const posY = Math.min(y1, y2) / moveSize;

    return `
    background-repeat: no-repeat;
    background-size: ${100 / size}%;
    background-position: ${posX * 100}% ${posY * 100}%;
    background-image: url(${avatar_url});
    `
}

async function displayUser() {
    if (loadingUser) {
        setTimeout(displayUser, 5)
        return;
    }

    if (!user) {
        document.getElementById("user-box").innerHTML = `<a id="user" class="login-account me-2" onclick="location.href='/login'" >Войти</a>`
        return;
    }

    const [[x1, y1], [x2, y2]] = user.avatar_box;
    const size = Math.abs(y1 - y2);
    const moveSize = 1 - size;
    const posX = Math.min(x1, x2) / moveSize;
    const posY = Math.min(y1, y2) / moveSize;

    let style =  `
    width: 1.5rem;
    height: 1.5rem;
    display: inline-block;
    `
    style += styleForAvatar(user.avatar_url, user.avatar_box)
    document.getElementById("user-box").innerHTML = `
    <div class="header-button">
    <a class="user-account" href="/users/${user.username}">
    <div class="user-avatar" style="${style}">
    </div>${user.username}</a>
    <div class="logout" onclick="logOut()">
    <svg xmlns="http://www.w3.org/2000/svg" width="21px" height="28px" fill="currentColor" class="bi bi-box-arrow-right" viewBox="0 0 16 16">
        <path fill-rule="evenodd" d="M10 12.5a.5.5 0 0 1-.5.5h-8a.5.5 0 0 1-.5-.5v-9a.5.5 0 0 1 .5-.5h8a.5.5 0 0 1 .5.5v2a.5.5 0 0 0 1 0v-2A1.5 1.5 0 0 0 9.5 2h-8A1.5 1.5 0 0 0 0 3.5v9A1.5 1.5 0 0 0 1.5 14h8a1.5 1.5 0 0 0 1.5-1.5v-2a.5.5 0 0 0-1 0z"/>
        <path fill-rule="evenodd" d="M15.854 8.354a.5.5 0 0 0 0-.708l-3-3a.5.5 0 0 0-.708.708L14.293 7.5H5.5a.5.5 0 0 0 0 1h8.793l-2.147 2.146a.5.5 0 0 0 .708.708z"/>
        </svg>
    </div>
    </div>`
}

async function bindPopups() {
    for (const elem of document.getElementsByClassName("small-popup")) {
        elem.addEventListener("mousedown", function (event) {
            if(event.target.classList.contains("small-popup")) {
                closePopup()
            }}
        );
    }
}

async function updateHash(hash) {
    const url = new URL(location);
    url.hash = hash
    history.pushState({}, "", url);
}

async function listenSearch() {
    const search = document.getElementById("search-box");

    search.addEventListener("focusin", (event) => {
        search.classList.add('active-search-box');
        search.focus()
    });

    search.addEventListener("focusout", (event) => {
        search.classList.remove('active-search-box');
    });

    const searchInput = document.getElementById('search-input');

    searchInput.addEventListener('keydown', function(event) {
        if (event.key === 'Enter') {
            event.preventDefault();
            location.href=`users/${searchInput.value}`;
        }
    });

    searchInput.addEventListener('search', function(event) {
        location.href=`users/${searchInput.value}`;
    });
}

document.addEventListener('scroll', () => {
    document.documentElement.style.setProperty('--scrollY', this.scrollY);
});

function onResize(event) {
    document.documentElement.style.setProperty('--screenX', window.innerWidth)
    document.documentElement.style.setProperty('--screenY', window.innerHeight)
}
addEventListener("resize", onResize, true);

onResize(null)