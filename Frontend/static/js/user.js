let userInfo, userRams, ramGenerator, ramPage;

const notFoundRam = new Error("Ram not found or ot not yours")

function show404() {
    document.getElementById("user-section").innerHTML = `<div class="text-center" style="margin-top: 32vh">
    <h1 class="m-1 mb-2">Такого пользователя не существует</h1>
    <a class="tap-text" style="font-size: 1.5rem" href="/">Вернуться на главную</a>
    </div>`;
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
            userInfo.own = (!!user && user.username === userInfo.username);
            return userInfo
        } else {
            show404();
        }
    } catch (error) {
        show404();
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
            const compareFn = (a, b) => a.id - b.id
            userRams.sort(compareFn)
            return userRams
        } else {
            show404();
        }
    } catch (error) {
        show404();
    }
}


let loadingUserInfo = true
let loadingUserRams = true
const userInfoUsername = window.location.pathname.split("/")[2];

loadUserInfo(userInfoUsername).then((user) => { if (user!==undefined) { loadingUserInfo = false }});
loadUserRams(userInfoUsername).then((rams) => { if (rams!==undefined) { loadingUserRams = false }});


async function displayUserInfo() {
    if (loadingUserInfo) {
        setTimeout(displayUserInfo, 5);
        return;
    }
    const [[x1, y1], [x2, y2]] = userInfo.avatar_box;
    const size = Math.abs(y1 - y2);
    let imageStyle =  `
    background-size: ${100 / size}%;
    background-position: ${(Math.min(y1, y2) + size / 2) * 100}% ${(Math.min(x1, x2) + size / 2) * 100}%;
    background-image: url(${userInfo.avatar_url});`;
    let imageOnclick = "";
    if (userInfo.avatar_ram_id !== 0) {
        imageStyle += "cursor: pointer;";
        imageOnclick = `onclick="location.hash='#ram-${userInfo.avatar_ram_id}'"`;
    }
    let res = `
    <div class="user-card-avatar" ${imageOnclick} style="${imageStyle}"></div>
    <h3 class="user-card-username">${userInfo.username}</h3>`;
    if (userInfo.own) {
        res += `<button class="button-user left-button-user row first-button-user" onclick="location.hash='#generate-ram'; ramGenerator = new Generator()">Сгенерировать барана</button>
                <button class="button-user left-button-user row " disabled onclick="location.href='/trade'">Обменять баранов</button>
                <button class="button-user left-button-user row last-button-user" onclick="location.hash='#settings'">Настройки аккаунта</button>`;
    } else {

    }
    document.getElementById("user-card").innerHTML = res;
}

function createRamElem(ram) {
    return `<div class="col">
        <div class="ram-card" onclick="openRam(${ram.id})">
          <img src="${ram.image_url}" class="ram-card-image" alt="${ram.description}">
          <h5 class="ram-card-description" style="text-align: center">${ram.description}</h5>
        </div>
      </div>`
}

async function displayUserRams() {
    if (loadingUserRams) {
        setTimeout(displayUserRams, 5);
        return;
    }
    if (userRams.length === 0) {
        if (userInfo.own) {
            document.getElementById("rams").innerHTML = `
            <div class="text-center top-to-center">
                <h2 class="m-1 mb-2">У вас пока-что нету баранов</h2>
                <a class="button-user" onclick="location.hash='#generate-ram';ramGenerator = new Generator()">Сгенерировать барана</a>
            </div>`;
        } else {
            document.getElementById("rams").innerHTML = `
            <div class="text-center top-to-center">
                <h2 class="m-1 mb-2">У этого пользователя нету баранов</h2>
            </div>`;
        }
        return
    }

    let res = "";
    for (let ram of userRams) {
        res += createRamElem(ram);
    }
    document.getElementById("rams").innerHTML = `<div id="rams-list" class="rams-list row row-cols-auto g-3 d-flex justify-content-start">${res}</div>`;
}

function appendRam(ram) {
    document.getElementById("rams-list").innerHTML += createRamElem(ram);
}

function countdown(target, elementId, callback) {
    const now = new Date().getTime();

    const distance = target - now;

    const hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
    const seconds = Math.floor((distance % (1000 * 60)) / 1000);

    let res = String(seconds).padStart(2, '0');
    if (minutes > 0 || hours > 0) {
        res = String(minutes).padStart(2, '0') + " : " + res;
    }
    if (hours > 0) {
        res = String(hours).padStart(2, '0') + " : " + res;
    }
    if (res[0] === "0") {
        res = res.slice(1);
    }
    document.getElementById(elementId).innerText = res;

    if (distance <= 0) {
        callback();
    }
}

class Clicker {
    constructor(clickElemId, onclickCallback, sendClicksCallback, start=0) {
        this.clickElemId = clickElemId;
        this.onclickCallback = onclickCallback;
        this.sendClicksCallback = sendClicksCallback;
        this.clicksCount = start;
        this.lastSendCount = start;
        this.clicksIferror = start;
        this.lastSendTime = new Date();
        this.onclickBinded = this.onclick.bind(this);

        document.getElementById(clickElemId).addEventListener("pointerup", this.onclickBinded);
        this.sendClicksInterval = setInterval(this.sendClicks.bind(this), 15000);
    }

    sendClicks(forced=false) {
        let now = new Date();

        if (!forced && (this.lastSendCount >= this.clicksCount || (+now - +this.lastSendTime) < 5 * 100)) {
            return
        }
        this.clicksIferror = this.clicksCount;
        this.sendClicksCallback(this.clicksCount-this.lastSendCount);
        this.lastSendCount = this.clicksCount;
        this.lastSendTime = now;
    }

    rollbackErrorClicks() {
        this.clicksCount = this.clicksIferror;
    }

    onclick() {
        this.clicksCount++;
        if (this.clicksCount - this.lastSendCount >= 50) {
            this.sendClicks();
        }
        this.onclickCallback(this.clicksCount);
    }

    close() {
        this.sendClicks(true);
        clearInterval(this.sendClicksInterval);
        document.getElementById(this.clickElemId).removeEventListener("pointerup", this.onclickBinded);
    }
}

class TargetedClicker extends Clicker {
    constructor(target, endCallback, clickElemId, onclickCallback, sendClicksCallback) {
        super(clickElemId, onclickCallback, sendClicksCallback);
        this.target = target;
        this.endCallback = endCallback;
    }

    onclick() {
        super.onclick();
        if (this.clicksCount >= this.target) {
            this.close();
            this.endCallback()
        }
    }
}

class Generator {
    constructor() {
        this.imageGenerated = false;
        this.initialize();
    }

    async initialize() {
        while (loadingUserInfo) {
            await sleep(5);
        }
        if (!userInfo.own) {
            location.hash = "";
        }

        document.querySelector("#generate-ram .popup-menu").innerHTML = `
             <h4 id="generation-title" class="text-center">Генерация барана</h2>
             <div id="generation-content" class="text-center"></div>`
        this.connectWs();
    }


    connectWs() {
        let apiUrl = new URL(API_URL);
        apiUrl.protocol = "ws";
        this.websocket = new WebSocket(`${apiUrl}/users/${user.username}/ws/generate-ram`);
        this.websocket.onopen = this._onopen.bind(this);
        this.websocket.onmessage = this._onmessage.bind(this);
        this.websocket.onclose = this._onclose.bind(this);
        this.websocket.onerror = this._onerror.bind(this);
    }

    sendPrompt() {
        const promptEl = document.getElementById("prompt");
        const prompt = promptEl.value;
        if (!prompt.trim().length) {
            return
        }
        this.websocket.send(prompt);
    }

    handleWSError(data) {
        let error;
        switch (data.code) {
            case 401:
                error = "Проблема с авторизацией, попробуйте выйти и зайти в аккаунт";
                break;
            case 403:
                error = "Вы не можете генерировать баранов для других пользователей";
                break
            case 409:
                error = "Похоже, вы тапаете барана в другой вкладке, или на другом устройстве. Попробуйте повторить позже.";
                break
            case 429:
                if (data.error.startsWith("you can generate only")) {
                    error = `Вы уже сгенерировали всех баранов на сегодня. Следующий раз вы сможете через <br><h1 id='error-countdown'></h1>`;
                    let cd = () => countdown(new Date(data.next * 1000), "error-countdown", () => {clearInterval(this.countdownInterval);ramGenerator = new Generator();});
                    setTimeout(cd, 10);
                    this.countdownInterval = setInterval(cd, 1000);
                } else {
                    error = `Нельзя так часто генерировать баранов! Вы сможете сгенерировать следующего барана через <br><h1 id='error-countdown'></h1>`;
                    let cd = () => countdown(new Date(data.next * 1000), "error-countdown", () => {clearInterval(this.countdownInterval);ramGenerator = new Generator();});
                    setTimeout(cd, 10);
                    this.countdownInterval = setInterval(cd, 1000);
                }
                break
            case 400:
                switch (data.error) {
                    case "invalid clicks":
                        this.targetedClicker.rollbackErrorClicks();
                        // TODO сообщение об ошибке
                        break;
                    default:
                        console.log("Unknown error", data.code, data.error);
                        error = `Unknown error ${data.code} ${data.error}`;
                        break;
                }
                break;
            //TODO либо закрывать popup, либо выводить сообщение об ошибке (с кнопкой ещё раз и без)
            default:
                console.log("Unknown error", data.code, data.error);
                error = `Unknown error ${data.code} ${data.error}`;
                break;
        }
        if (error) {
            const content = document.getElementById("generation-content");
            content.innerHTML = `
                <div class="text-center popup-error" style="position: relative;top: 30%">
                    <h5>${error}</h5>
                    <a class="tap-text" style="font-size: 1rem" onclick="closePopup()">ОК</a>
                </div>
                `;
        }
    }

    close() {
        if (this.websocket.readyState !== WebSocket.CLOSED) {
            if (this.targetedClicker) {
                this.targetedClicker.close();
            }
            this.websocket.close();
        }

        if (this.countdownInterval) {
            clearInterval(this.countdownInterval);
        }

        document.getElementById("generation-content").innerHTML = ``;
        ramGenerator = undefined;
    }

    endClicker () {
        this.targetedClicker = undefined;
    }

    changeClicksValue(value) {
        const clicksEl = document.getElementById("clicks");
        clicksEl.innerText = `${value}/${this.needClicks}`;
    }

    sendClicks(value) {
        console.log(value)
        this.websocket.send(`${value}`);
    }

    _onopen(event) {
        this.websocket.send(getCookie("token"));
    }

    _onmessage(event) {
        let data = JSON.parse(event.data);
        console.log("Message: ", data);
        if (data.error) {
            this.handleWSError(data);
            return
        }
        if (data.id) {
            this.close();
            // TODO анимация завершения
            appendRam(data)
            openRam(data.id);
            return
        }
        const content = document.getElementById("generation-content");
        switch (data.status) {
            case "need first ram prompt":
                content.innerHTML = `
                <label class="mb-4 prompt-label">
                    Введите запрос для вашего первого барана<br>
                    <input class="prompt-input" id="prompt" type="text" maxlength="30">
                </label>
                <div id="generation-bottom" class="text-center mt-auto">
                    <button id="enter-prompt" class="button-user" onclick="ramGenerator.sendPrompt()">Далее</button>
                </div>`;

                break;
            case "need ram prompt":
                content.innerHTML = `
                <label class="mb-4 prompt-label">
                    Введите запрос для барана<br>
                    <input class="prompt-input" id="prompt" type="text" maxlength="30">
                </label>
                <div id="generation-bottom" class="text-center mt-auto">
                    <button id="enter-prompt" class="button-user" onclick="ramGenerator.sendPrompt()">Далее</button>
                </div>`;
                break;
            case "need clicks":
                this.needClicks = data.clicks;
                content.innerHTML = `
                    <h4 class="text-center tap-label">Тапните ${this.needClicks} раз, чтобы сгенерировать барана</h4>
                    <img id="clicker" src="/static/img/rambox1.png" class="tap-generate-img" alt="" style="cursor: pointer;">
                    <h2 id="clicks" class="text-center">0/${this.needClicks}</h2>`;
                this.targetedClicker = new TargetedClicker(
                    this.needClicks,
                    this.endClicker.bind(this),
                    "clicker",
                    this.changeClicksValue.bind(this),
                    this.sendClicks.bind(this));
                break;
            case "image generated":
                this.imageGenerated = true;
                break;
            case "success clicked":
                if (this.imageGenerated) {
                    // TODO анимация завершения
                } else {
                    content.innerHTML = `
                    <h4 class="text-center" style="margin-top: 20%">Вы очень быстро тапали!</h4>
                    <h6 class="text-center">Подождите, баран ещё не успел сгенерироваться...</h6>
                    <img id="wait-ram" src="/static/img/icon512.png" class="img-fluid wait-ram" alt="" style="cursor: pointer;">
                    <h3 id="wait-clicks" class="text-center"></h3>`;
                    this.targetedClicker = new Clicker(
                        "wait-ram",
                        function (value) {
                            const clicksEl = document.getElementById("wait-clicks");
                            clicksEl.innerText = `${value}`;
                        }.bind(this),
                        function () {});
                }
                break;
        }
    }


    _onclose(event) {
        if (event.wasClean) {
            console.log(`[close] Соединение закрыто чисто, код=${event.code} причина=${event.reason}`);
        } else {
            let reason;
            switch (event.code) {
                // TODO
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
        const content = document.getElementById("generation-content");
        content.innerHTML = `
                <div class="text-center popup-error" style="margin-top: 40%">
                    <h4>Не удалось подключиться к серверу.</h4>
                    <a class="tap-text" style="font-size: 1rem" onclick="closePopup()">ОК</a>
                </div>`;
    };
}

async function getRam(username, id) {
    const response = await fetch(`${API_URL}/users/${username}/rams/${id}`, {
        mode: 'cors',
        method: 'GET',
    });
    if (response.ok) {
        const ram = await response.json();
        console.log(user, ram)
        ram.own = (!!user && user.id === ram.user_id);
        return ram
    } else {
        throw notFoundRam;
    }
}

class RamPage {
    constructor(id) {
        getRam(userInfoUsername, id).then(
            ram => {
                this.ram = ram;
                document.querySelector("#ram .popup-menu").innerHTML = `
             <h4 id="ram-description" class="text-center">${this.ram.description}</h2>
             <div id="ram-content" class="text-center">
                <img id="ram-clicker" class="ram-image mt-5" src="${this.ram.image_url}" alt="ram">
                <h3 id="ram-clicked" class="mt-3">${this.ram.taps} тапов</h3>
             </div>`;
                if (this.ram.own) {
                    document.getElementById("ram-clicker").classList.add("cursor-pointer");
                    this.clicker = new Clicker(
                        "ram-clicker",
                        function (value) { document.getElementById("ram-clicked").innerHTML = `${value} тапов`},
                        this.sendClicks.bind(this),
                        this.ram.taps,
                    );
                    this.connectWs();
                }
            },
            error => {
                console.log(error);
                document.querySelector("#ram .popup-menu").innerHTML = `
                <div class="text-center popup-error ram-error">
                    <h5>Такого барана не найдено.</h5>
                    <a class="tap-text" style="font-size: 1rem" onclick="closePopup()">ОК</a>
                </div>`;
            }
        );
    }

    connectWs() {
        let apiUrl = new URL(API_URL);
        apiUrl.protocol = "ws";
        this.websocket = new WebSocket(`${apiUrl}/users/${user.username}/rams/${this.ram.id}/ws/clicker`);
        this.websocket.onopen = this._onopen.bind(this);
        this.websocket.onmessage = this._onmessage.bind(this);
        this.websocket.onclose = this._onclose.bind(this);
        this.websocket.onerror = this._onerror.bind(this);
    }

    handleWSError(data) {
        let error;
        switch (data.code) {
            case 401:
                error = "Проблема с авторизацией, попробуйте выйти и зайти в аккаунт";
                break;
            case 403:
                error = "Вы не можете тапать баранов для других пользователей";
                break
            case 409:
                error = "Похоже, вы тапаете барана в другой вкладке, или на другом устройстве. Попробуйте повторить позже.";
                break
            case 400:
                switch (data.error) {
                    case "invalid clicks":
                        this.targetedClicker.rollbackErrorClicks();
                        // TODO сообщение об ошибке
                        break;
                    default:
                        console.log("Unknown error", data.code, data.error);
                        error = `Unknown error ${data.code} ${data.error}`;
                        break;
                }
                break;
            //TODO либо закрывать popup, либо выводить сообщение об ошибке (с кнопкой ещё раз и без)
            default:
                console.log("Неизвестная ошибка", data.code, data.error);
                error = `Неизвестная ошибка ${data.code} ${data.error}`;
                break;
        }
        if (error) {
            const content = document.getElementById("ram-content");
            content.innerHTML = `
                <div class="text-center popup-error" style="position: relative;top: 30%">
                    <h5>${error}</h5>
                    <a class="tap-text" style="font-size: 1rem" onclick="closePopup()">ОК</a>
                </div>
                `;
        }
    }

    close() {
        if (this.ram && this.ram.own) {
            if (this.websocket.readyState !== WebSocket.CLOSED) {
                if (this.clicker) {
                    this.clicker.close();
                }
                this.websocket.close();
            }
        }

        ramPage = undefined;
        document.querySelector("#ram .popup-menu").innerHTML = ``;
    }

    sendClicks(value) {
        console.log(value)
        this.websocket.send(`${value}`);
    }

    _onopen(event) {
        this.websocket.send(getCookie("token"));
    }

    _onmessage(event) {
        let data = JSON.parse(event.data);
        console.log("Message: ", data);
        if (data.error) {
            this.handleWSError(data);
        }
    }

    _onclose(event) {
        if (event.wasClean) {
            console.log(`[close] Соединение закрыто чисто, код=${event.code} причина=${event.reason}`);
        } else {
            let reason;
            switch (event.code) {
                // TODO
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
        const content = document.getElementById("ram-content");
        content.innerHTML = `
                <div class="text-center popup-error" style="margin-top: 40%">
                    <h4>Не удалось подключиться к серверу.</h4>
                    <a class="tap-text" style="font-size: 1rem" onclick="closePopup()">ОК</a>
                </div>`;
    };
}

async function openRam(id) {
    location.hash = "#ram";

    const url = new URL(location);
    url.searchParams.set("ram-id", `${id}`);
    history.pushState({}, "", url);
    ramPage = new RamPage(id)
}

function bindUserPopups() {
    for (const elem of document.getElementsByClassName("user-popup")) {
        elem.addEventListener("mousedown", function (event) {
            if(event.target.classList.contains("user-popup")) {
                closePopup()
            }}
        );
    }
}

function checkHash() {
    if (location.hash === "#generate-ram" && user.username === userInfoUsername) {
        ramGenerator = new Generator();
    }
    if (location.hash === "#ram") {
        const url = new URL(location);
        ramPage = new RamPage(parseInt(url.searchParams.get("ram-id")))
    }
}


function closePopup() {
    if (ramGenerator) {
        ramGenerator.close();
    }
    if (ramPage) {
        ramPage.close();
    }
    location.hash = "";

    const url = new URL(location);
    if (url.searchParams.has("ram-id")) {
        url.searchParams.delete("ram-id");
        history.pushState({}, "", url);
    }
}

window.addEventListener('beforeunload', closePopup);
