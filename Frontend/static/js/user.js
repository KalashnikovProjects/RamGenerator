let userInfo, userRams, ramGenerator, needClicks, clicks;

function show404() {
    document.getElementById("user-section").innerHTML = `<div class="text-center" style="margin-top: 32vh">
    <h1 class="m-1 mb-2">Такого пользователя не существует</h1>
    <a class="tap-text" style="font-size: 1.5rem" href="/">Вернуться на главную</a>
    </div>`
}

async function loadUserInfo(username) {
    try {
        const response = await fetch(`${API_URL}/users/${username}`, {
            mode: 'cors',
            method: 'GET',
        });
        if (response.ok) {
            userInfo = await response.json();
            userInfo.avatar_url = await loadAvatar(userInfo.username, userInfo.avatar_ram_id);
            userInfo.own = (!!user && user.username === userInfo.username)
            return userInfo
        } else {
            show404()
        }
    } catch (error) {
        show404()
    }
}


async function loadUserRams(username) {
    try {
        const response = await fetch(`${API_URL}/users/${username}/rams`, {
            mode: 'cors',
            method: 'GET',
        });
        if (response.ok) {
            userRams = await response.json();
            return userRams
        } else {
            show404()
        }
    } catch (error) {
        show404()
    }
}


let loadingUserInfo = true
let loadingUserRams = true
const userInfoUsername = window.location.pathname.split("/")[2]

loadUserInfo(userInfoUsername).then((user) => { if (user!==undefined) { loadingUserInfo = false }})
loadUserRams(userInfoUsername).then((rams) => { if (rams!==undefined) { loadingUserRams = false }})


async function displayUserInfo() {
    if (loadingUserInfo) {
        setTimeout(displayUserInfo, 5)
        return;
    }
    const [[x1, y1], [x2, y2]] = userInfo.avatar_box;
    const size = Math.abs(y1 - y2)
    let imageStyle =  `
    background-size: ${100 / size}%;
    background-position: ${(Math.min(y1, y2) + size / 2) * 100}% ${(Math.min(x1, x2) + size / 2) * 100}%;
    background-image: url(${userInfo.avatar_url});`
    let imageOnclick = ""
    if (userInfo.avatar_ram_id !== 0) {
        imageStyle += "cursor: pointer;"
        imageOnclick = `onclick="location.hash='#ram-${userInfo.avatar_ram_id}'"`
    }
    let res = `
    <div class="user-card-avatar" ${imageOnclick} style="${imageStyle}"></div>
    <h3 class="user-card-username">${userInfo.username}</h3>`
    if (userInfo.own) {
        res += `<button class="button-user left-button-user row first-button-user" onclick="location.hash='#generate-ram'; ramGenerator = new Generator()">Сгенерировать барана</button>
                <button class="button-user left-button-user row " disabled onclick="location.href='/trade'">Обменять баранов</button>
                <button class="button-user left-button-user row last-button-user" onclick="location.hash='#settings'">Настройки аккаунта</button>`
    } else {

    }
    document.getElementById("user-card").innerHTML = res
}

async function displayUserRams() {
    if (loadingUserRams) {
        setTimeout(displayUserRams, 5)
        return;
    }
    if (userRams.length === 0) {
        if (userInfo.own) {
            document.getElementById("rams").innerHTML = `
            <div class="text-center top-to-center">
                <h2 class="m-1 mb-2">У вас пока-что нету баранов</h2>
                <a class="button-user" onclick="location.hash='#generate-ram';ramGenerator = new Generator()">Сгенерировать барана</a>
            </div>`
        } else {
            document.getElementById("rams").innerHTML = `
            <div class="text-center top-to-center">
                <h2 class="m-1 mb-2">У этого пользователя нету баранов</h2>
            </div>`
        }
        return
    }
    document.getElementById("rams").innerHTML = `<div class="rams-list row row-cols-auto g-3 d-flex justify-content-center"></div>`
    for (let ram in userRams) {
        // TODO
    }
}

class Clicker {
    constructor(clickElemId, sendClicksCallback) {
        this.clickElemId = clickElemId
        this.sendClicksCallback = sendClicksCallback

        this.clicksCount = 0
        this.lastSendCount = 0
        this.lastSendTime = new Date();

        document.getElementById(clickElemId).addEventListener("pointerup", this.onclick.bind(this))
    }

    onclick() {
        this.clicksCount++
    }

    close() {

    }
}

class Generator {
    constructor() {
        this.initialize()
    }

    async initialize() {
        while (loadingUserInfo) {
            await sleep(5)
        }
        if (!userInfo.own) {
            location.hash = ""
        }

        document.querySelector("#generate-ram .popup-menu").innerHTML = `
             <h4 id="generation-title" class="text-center">Генерация барана</h2>
             <div id="generation-content"></div>`
        this.connectWs()
    }


    connectWs() {
        let apiUrl = new URL(API_URL)
        apiUrl.protocol = "ws"
        this.websocket = new WebSocket(`${apiUrl}/users/${user.username}/ws/generate-ram`);
        this.websocket.onopen = this._onopen.bind(this)
        this.websocket.onmessage = this._onmessage.bind(this)
        this.websocket.onclose = this._onclose.bind(this)
        this.websocket.onerror = this._onerror.bind(this)
    }

    sendPrompt() {
        const promptEl = document.getElementById("prompt")
        const prompt = promptEl.value
        if (!prompt.trim().length) {
            // TODO: error
            return
        }
        this.websocket.send(prompt);
    }

    handleWSError(data) {
        let error;
        switch (data.code) {
            case 401:
                error = "Проблема с авторизацией, попробуйте выйти и зайти в аккаунт"
                break;
            case 403:
                error = "Вы не можете генерировать баранов для других пользователей"
                break
            case 409:
                error = "Похоже, вы тапаете барана в другой вкладке, или на другом устройстве. Попробуйте повторить позже."
                break
            //TODO
            default:
                console.log("Unknown error", data.error);
                error = `Unknown error ${data.error}`
                break;
        }
        if (error) {
            const content = document.getElementById("generation-content")
            content.innerHTML = `
                <div class="text-center popup-error" style="position: relative;top: 30%">
                    <h5>${error}</h5>
                    <a class="tap-text" style="font-size: 1rem" onclick="closePopup()">ОК</a>
                </div>
                `
        }
    }

    close() {
        document.getElementById("generation-content").innerHTML = ``
        this.websocket.close()
        needClicks = undefined
        clicks = 0
        ramGenerator = undefined
    }

    _onopen(event) {
        this.websocket.send(getCookie("token"));
    }

    _onmessage(event) {
        let data = JSON.parse(event.data)
        console.log("Message: ", data)
        if (data.error) {
            this.handleWSError(data)
            return
        }
        const content = document.getElementById("generation-content");
        switch (data.status) {
            case "need first ram prompt":
                content.innerHTML = `
                <label class="mb-4 prompt-label">
                    Введите запрос для вашего первого барана<br>
                    <input class="prompt-input" id="prompt" type="text" name="" placeholder="" maxlength="30">
                </label>
                <div id="generation-bottom">
                    <button id="enter-prompt" class="button-user" onclick="ramGenerator.sendPrompt()">Далее</button>
                </div>`

                break;
            case "need ram prompt":
                content.innerHTML = `
                <label class="mb-4 prompt-label">
                    Введите запрос для барана<br>
                    <input class="prompt-input" id="prompt" type="text" name="" placeholder="" maxlength="30">
                </label>
                <div id="generation-bottom">
                    <button id="enter-prompt" class="button-user" onclick="ramGenerator.sendPrompt()">Далее</button>
                </div>`
                break;
            case "need clicks":
                needClicks = data.clicks
                clicks = 0
                content.innerHTML = `
                    <h4 class="text-center tap-label">Тапните ${needClicks} раз, чтобы сгенерировать барана</h4>
                    <img src="/static/img/rambox1.png" class="img-fluid tap-generate-img" alt="" style="cursor: pointer;">
                    <h1 id="clicks" class="text-center">${clicks}/${needClicks}</h1>`
                //TODO
                break;
            //TODO
        }
    }

    _onclose(event) {
        if (event.wasClean) {
            console.log(`[close] Соединение закрыто чисто, код=${event.code} причина=${event.reason}`);
        } else {
            let reason;
            switch (event.code) {
                case 1000:
                    reason = "Normal closure, meaning that the purpose for which the connection was established has been fulfilled.";
                    break;
                case 1001:
                    reason = "An endpoint is \"going away\", such as a server going down or a browser having navigated away from a page.";
                    break;
                case 1002:
                    reason = "An endpoint is terminating the connection due to a protocol error";
                    break;
                case 1003:
                    reason = "An endpoint is terminating the connection because it has received a type of data it cannot accept (e.g., an endpoint that understands only text data MAY send this if it receives a binary message).";
                    break;
                case 1004:
                    reason = "Reserved. The specific meaning might be defined in the future.";
                    break
                case 1005:
                    reason = "No status code was actually present.";
                    break
                case 1006:
                    reason = "The connection was closed abnormally, e.g., without sending or receiving a Close control frame";
                    break
                case 1007:
                    reason = "An endpoint is terminating the connection because it has received data within a message that was not consistent with the type of the message (e.g., non-UTF-8 [https://www.rfc-editor.org/rfc/rfc3629] data within a text message).";
                    break
                case 1008:
                    reason = "An endpoint is terminating the connection because it has received a message that \"violates its policy\". This reason is given either if there is no other sutible reason, or if there is a need to hide specific details about the policy.";
                    break
                case 1009:
                    reason = "An endpoint is terminating the connection because it has received a message that is too big for it to process.";
                    break
                case 1010:
                    reason = "An endpoint (client) is terminating the connection because it has expected the server to negotiate one or more extension, but the server didn't return them in the response message of the WebSocket handshake. Specifically, the extensions that are needed are: " + event.reason;
                    break
                case 1011:
                    reason = "A server is terminating the connection because it encountered an unexpected condition that prevented it from fulfilling the request.";
                    break
                case 1015:
                    reason = "The connection was closed due to a failure to perform a TLS handshake (e.g., the server certificate can't be verified).";
                    break
                default:
                    reason = "Unknown reason";
            }
        }
    }

    _onerror(error) {
        const content = document.getElementById("generation-content")
        content.innerHTML = `
                <div class="text-center popup-error" style="margin-top: 40%">
                    <h4>Ошибка подключения, возможно вы открываете эту страницу слишком часто.</h4>
                    <a class="tap-text" style="font-size: 1rem" onclick="closePopup()">ОК</a>
                </div>`
    };
}

function bindUserPopups() {
    for (const elem of document.getElementsByClassName("user-popup")) {
        elem.addEventListener("mousedown", function (event) {
            if(event.target.classList.contains("user-popup")) {closePopup()}})
    }
}

function checkHash() {
    if (location.hash === "#generate-ram" && user.username === userInfoUsername) {
        ramGenerator = new Generator()
    }
}


function closePopup() {
    if (ramGenerator) {
        ramGenerator.close()
    }
    location.hash = ""
}
window.addEventListener('beforeunload', closePopup);