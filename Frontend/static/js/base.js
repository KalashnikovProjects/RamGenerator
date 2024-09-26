"use strict";

let user;

function sleep(time) {
    return new Promise((resolve) => setTimeout(resolve, time));
}

if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/static/js/sw.js')
        .then(() => navigator.serviceWorker.ready.then((worker) => {
            worker.sync.register('syncdata');
        }))
        .catch((err) => console.log(err));
}

async function loadAvatar(username, ramId) {
    if (ramId === 0) {
        return DEFAULT_AVATAR
    }
    const response = await fetch(`${API_URL}/users/${username}/rams/${ramId}`, {
        mode: 'cors',
        method: 'GET',
    });

    if (response.ok) {
        let ram = await response.json();
        return ram.image_url
    } else {
        const errorText = await response.text();
        console.error('Error response:', response.status, errorText);
        return DEFAULT_AVATAR
    }
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
            user.avatar_url = await loadAvatar(user.username, user.avatar_ram_id);
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

    const style =  `
    width: 1.5rem;
    height: 1.5rem;
    background-repeat: no-repeat;
    display: inline-block;
    background-size: ${100 / size}%;
    background-position: ${posX * 100}% ${posY * 100}%;
    background-image: url(${user.avatar_url});
    `
    document.getElementById("user-box").innerHTML = `
    <div class="header-button">
    <a class="user-account" href="/users/${user.username}">
    <div class="user-avatar" style="${style}">
    </div>${user.username}</a>
    <div class="logout" onclick="logOut()">
    <svg xmlns="http://www.w3.org/2000/svg" width="1.2rem" height="1.2rem" fill="currentColor" class="bi bi-box-arrow-right" viewBox="0 0 16 16">
        <path fill-rule="evenodd" d="M10 12.5a.5.5 0 0 1-.5.5h-8a.5.5 0 0 1-.5-.5v-9a.5.5 0 0 1 .5-.5h8a.5.5 0 0 1 .5.5v2a.5.5 0 0 0 1 0v-2A1.5 1.5 0 0 0 9.5 2h-8A1.5 1.5 0 0 0 0 3.5v9A1.5 1.5 0 0 0 1.5 14h8a1.5 1.5 0 0 0 1.5-1.5v-2a.5.5 0 0 0-1 0z"/>
        <path fill-rule="evenodd" d="M15.854 8.354a.5.5 0 0 0 0-.708l-3-3a.5.5 0 0 0-.708.708L14.293 7.5H5.5a.5.5 0 0 0 0 1h8.793l-2.147 2.146a.5.5 0 0 0 .708.708z"/>
        </svg>
    </div>
    </div>`
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
