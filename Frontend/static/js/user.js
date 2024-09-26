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
    const moveSize = 1 - size;
    const posX = Math.min(x1, x2) / moveSize;
    const posY = Math.min(y1, y2) / moveSize;
    let imageStyle =  `
    background-size: ${100 / size}%;
    background-position: ${posX * 100}% ${posY * 100}%;
    background-image: url(${userInfo.avatar_url});`;
    let imageOnclick = "";
    if (userInfo.avatar_ram_id !== 0) {
        imageStyle += "cursor: pointer;";
        imageOnclick = `onclick="openRam(${userInfo.avatar_ram_id})"`;
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
    let el = document.getElementById("rams-list");
    let elem = createRamElem(ram);
    if (!el) {
        document.getElementById("rams").innerHTML = `<div id="rams-list" class="rams-list row row-cols-auto g-3 d-flex justify-content-start">${elem}</div>`;
    } else {
        el.innerHTML += elem
    }
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
        try {
            this.sendClicks(true);
        } catch (e) {}
        if (this.sendClicksInterval) {
            clearInterval(this.sendClicksInterval);
        }
        try {
            document.getElementById(this.clickElemId).removeEventListener("pointerup", this.onclickBinded);
        } catch (e) {}

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
        document.querySelector("#generate-ram .popup-menu").innerHTML = `
             <h4 id="generation-title" class="text-center">Генерация барана</h2>
             <div id="generation-content" class="text-center">
                <img id="loading-image-generator" src="/static/img/icon512.png" class="loading-image rotating-image img-fluid wait-ram" alt="Загрузка..." style="cursor: pointer">
            </div>
             <button id="close-button" style="right:1.5rem" class="up-button" onclick="closePopup()">
                 <svg xmlns="http://www.w3.org/2000/svg" fill="white" class="bi bi-x" viewBox="0 0 16 16">
                    <path d="M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708"/>
                </svg>
            </button>`
        setTimeout(() => document.getElementById("loading-image-generator").classList.add("loading-image-visible"), 10)
        while (loadingUserInfo) {
            await sleep(5);
        }
        if (!userInfo.own) {
            location.hash = "";
        }
        this.connectWs();
    }


    connectWs() {
        let apiUrl = new URL(API_URL);
        apiUrl.protocol = WEBSOCKET_PROTOCOL;
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
                    let cd = () => countdown(new Date(data.next * 1000), "error-countdown", () => {clearInterval(this.countdownInterval);this.close();ramGenerator = new Generator();});
                    setTimeout(cd, 10);
                    this.countdownInterval = setInterval(cd, 1000);
                } else {
                    error = `Нельзя так часто генерировать баранов! Вы сможете сгенерировать следующего барана через <br><h1 id='error-countdown'></h1>`;
                    let cd = () => countdown(new Date(data.next * 1000), "error-countdown", () => {clearInterval(this.countdownInterval);this.close();ramGenerator = new Generator();});
                    setTimeout(cd, 10);
                    this.countdownInterval = setInterval(cd, 1000);
                }
                break
            case 400:
                switch (data.error) {
                    case "invalid clicks":
                        try {
                            this.targetedClicker.rollbackErrorClicks();
                        } catch (e) {}
                        // TODO сообщение об ошибке
                        break;
                    case "user prompt or rams descriptions contains illegal content":
                        error = `Не получилось сгенерировать барана по такому запросу, попробуйте ещё раз`;
                        break;
                    default:
                        console.log("Unknown error", data.code, data.error);
                        error = `Unknown error ${data.code} ${data.error}`;
                        break;
                }
                break;
            case 500:
                switch (data.error) {
                    case "read message error":
                    case "send message error":
                        break;
                    case "image generation timeout":
                        error = `Сервис генерации изображений баранов не отвечает, попробуйте позже`;
                        break;
                    case "image generating error":
                    case "image generation service unavailable":
                        error = `Сервис генерации изображений баранов сейчас недоступен, попробуйте позже`;
                        break;
                    case "image description generating error":
                    case "image uploading error":
                    case "prompt generating error":
                        error = `Сервис, необходимый для генерации баранов сейчас недоступен, попробуйте позже`;
                        break;
                    default:
                        error = "Произошла неизвестная ошибка на стороне сервера.";
                        console.log("Unknown internal server error", data.code, data.error);
                        break;
                }
                break
            //TODO при других ошибках либо закрывать popup, либо выводить сообщение об ошибке (с кнопкой ещё раз или без)
            default:
                console.log("Unknown error", data.code, data.error);
                error = `Произошла неизвестная ошибка`;
                break;
        }
        if (error && !this.preventError) {
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
        this.preventError = true
        if (this.websocket.readyState !== WebSocket.CLOSED) {
            try {
                if (this.targetedClicker) {
                    this.targetedClicker.close();
                }
                this.websocket.close();
            } catch (e) {}
        }

        if (this.countdownInterval) {
            clearInterval(this.countdownInterval);
        }
        try {
            document.getElementById("generation-content").innerHTML = ``;
        } catch (e) {}
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
            appendRam(data);
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

    _onclose(event) {}

    _onerror(error) {
        if (!this.preventError) {
            const content = document.getElementById("generation-content");
            content.innerHTML = `
                <div class="text-center popup-error" style="margin-top: 40%">
                    <h4>Не удалось подключиться к серверу.</h4>
                    <a class="tap-text" style="font-size: 1rem" onclick="closePopup()">ОК</a>
                </div>`;
        }
    };
}

async function getRam(username, id) {
    const response = await fetch(`${API_URL}/users/${username}/rams/${id}`, {
        mode: 'cors',
        method: 'GET',
    });
    if (response.ok) {
        const ram = await response.json();
        ram.own = (!!user && user.id === ram.user_id);
        return ram
    } else {
        throw notFoundRam;
    }
}

class RamPage {
    constructor(id) {
        let elem =  document.querySelector("#ram .popup-menu");
        elem.innerHTML = `
             <h4 id="ram-description" class="ram-description">Загрузка барана...</h2>
             <img id="loading-image-ram-page" src="/static/img/icon512.png" class="loading-image rotating-image img-fluid wait-ram" alt="Загрузка..." style="cursor: pointer">
             <button id="close-button" style="right:1.5rem" class="up-button" onclick="closePopup()">
                 <svg xmlns="http://www.w3.org/2000/svg" fill="white" class="bi bi-x" viewBox="0 0 16 16">
                    <path d="M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708"/>
                </svg>
            </button>
            `;
        setTimeout(() => document.getElementById("loading-image-ram-page").classList.add("loading-image-visible"), 10)
        getRam(userInfoUsername, id).then(
            ram => {
                this.ram = ram;
                let elem =  document.querySelector("#ram .popup-menu");
                elem.innerHTML = `
             <h4 id="ram-description" class="ram-description">${this.ram.description}</h2>
             <div id="ram-content" class="text-center ram-content">
                <img id="ram-clicker" class="ram-image mt-5" src="${this.ram.image_url}" alt="ram">
                <div id="taps-line" class="mt-3"><h3 id="ram-clicked">${this.ram.taps} тапов</h3></div>
             </div>
             <button id="close-button" style="right:1.5rem" class="up-button" onclick="closePopup()">
                 <svg xmlns="http://www.w3.org/2000/svg" fill="white" class="bi bi-x" viewBox="0 0 16 16">
                    <path d="M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708"/>
                </svg>
            </button>
            `;
                if (this.ram.own) {
                    document.getElementById("ram-clicker").classList.add("cursor-pointer");
                    elem.innerHTML += `<button id="avatar-button" style="left:1.5rem" class="up-button" onclick="toggleCropMode()">
                                            <img src="/static/img/avatar-button.svg" alt="На аватар">
                                        </button>`;
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
        apiUrl.protocol = WEBSOCKET_PROTOCOL;
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
                        try {
                            this.targetedClicker.rollbackErrorClicks();
                        } catch (e) {}
                        // TODO сообщение об ошибке
                        break;
                    default:
                        console.log("Unknown error", data.code, data.error);
                        error = `Unknown error ${data.code} ${data.error}`;
                        break;
                }
                break;
            //TODO при других ошибках либо закрывать popup, либо выводить сообщение об ошибке (с кнопкой ещё раз и без)
            default:
                console.log("Неизвестная ошибка", data.code, data.error);
                error = `Неизвестная ошибка ${data.code} ${data.error}`;
                break;
        }
        if (error && !this.preventError) {
            const content = document.getElementById("ram-content");
            content.innerHTML = `
                <div class="text-center popup-error" style="position: relative;top: 30%">
                    <h5>${error}</h5>
                    <a class="tap-text" style="font-size: 1rem" onclick="closePopup()">ОК</a>
                </div>
                `;
        }
    }

    close(destroy=true) {
        this.preventError = true
        if (this.websocket.readyState !== WebSocket.CLOSED) {
            try {
                if (this.clicker) {
                    this.clicker.close();
                }
                this.websocket.close();
            } catch (e) {}
        }
        if (destroy) {
            try {
                document.querySelector("#ram .popup-menu").innerHTML = ``
            } catch (e) {}
        }
        ramPage = undefined;
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

    _onclose(event) {}

    _onerror(error) {
        if (!this.preventError) {
            const content = document.getElementById("ram-content");
            content.innerHTML = `
                <div class="text-center popup-error" style="margin-top: 40%">
                    <h4>Не удалось подключиться к серверу.</h4>
                    <a class="tap-text" style="font-size: 1rem" onclick="closePopup()">ОК</a>
                </div>`;
        }
    }
}

async function openRam(id) {
    location.hash = "#ram";

    const url = new URL(location);
    url.searchParams.set("ram-id", `${id}`);
    history.pushState({}, "", url);
    ramPage = new RamPage(id)
}

function validateUsername(username) {
    return username.length >= 3 && username.length <= 24;
}

function validatePassword(password, passwordRepeat = null) {
    if (password === "") {
        return "Необходимо заполнить поле пароля";
    }
    if (passwordRepeat !== null && password !== passwordRepeat) {
        return "Пароли не совпадают";
    }
    return null;
}

async function responseProcess(el, response, okCallback=null) {
    if (response.ok) {
        if ( !!okCallback && typeof okCallback === "function") {
            okCallback();
        }
    } else {
        el.classList.add("text-danger");
        const text = await response.text();

        let errorText;
        switch (response.status) {
            case 404:
                errorText = "Такого пользователя не существует";
                break;
            case 401:
                errorText = "Вы не можете редактировать чужой профиль, возможно вы не вошли в аккаунт";
                break;
            case 400:
                if (text.startsWith("username must be ")) {
                    errorText = "Имя должно состоять только из английских букв, цифр и нижнего подчёркивания";
                } else if (text.startsWith("required fields are not specified")) {
                    errorText = `Не заполнены необходимые поля: ${text.split(": ")[1]}`;
                } else if (text.includes("is already taken")) {
                    errorText = `Имя ${text.split("username ")[1].split(" is already taken")[0]} уже занято`;
                } else {
                    errorText = `Не заполнены необходимые поля: ${text.split(": ")[1]}`;
                }
                break;
            default:
                errorText = "Произошла неизвестная ошибка на сервере, попробуйте войти позже";
        }
        el.innerText = errorText
    }
}


function changeUsername(event) {
    event.preventDefault();
    const el = document.getElementById('username-message')

    const username = document.getElementById('settings-username').value;

    if (!validateUsername(username)) {
        el.classList.add("text-danger");
        el.innerText = "Имя должно содержать от 3 до 24 символов";
        return;
    }

    fetch(`${API_URL}/users/${userInfoUsername}`, {
        mode: 'cors',
        method: 'PATCH',
        headers: {
            'Authorization': `Bearer ${getCookie("token")}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({"username": username})
    }).then(
        (response) => {responseProcess(el, response, () => {
            sessionStorage.removeItem("user");
            loadUser().then(() => {
                el.classList.remove("text-danger");
                el.innerText = "Успешно сохранено, сейчас страница перезагрузится"
                setTimeout(() => {
                    location.hash = "";
                    location.href = `users/${user.username}`;
                }, 3000)
            });
        })},
        (error) => {
            el.classList.add("text-danger");
            el.innerText = "Произошла ошибка при сохранении"
        });
}

function changePassword(event) {
    event.preventDefault();
    const el = document.getElementById('password-message')
    const password = document.getElementById('settings-password').value;
    const passwordRepeat = document.getElementById('settings-password-repeat').value;
    const err = validatePassword(password, passwordRepeat);
    if (err) {
        el.classList.add("text-danger");
        el.innerText = err;
        return;
    }
    fetch(`${API_URL}/users/${userInfoUsername}`, {
        mode: 'cors',
        method: 'PATCH',
        headers: {
            'Authorization': `Bearer ${getCookie("token")}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({"password": password})
    }).then(
        (response) => {responseProcess(el, response, () => {
            el.classList.remove("text-danger");
            el.innerText = "Успешно сохранено";
        })},
        (error) => {
            el.classList.add("text-danger");
            el.innerText = "Произошла ошибка при сохранении"
        });
}

async function bindSettingsForms() {
    const changeUsernameForm = document.getElementById('changeUsernameForm');
    if (changeUsernameForm) {
        changeUsernameForm.addEventListener('submit', changeUsername);
    }

    const changePasswordForm = document.getElementById('changePasswordForm');
    if (changePasswordForm) {
        changePasswordForm.addEventListener('submit', changePassword);
    }
}

async function bindUserPopups() {
    for (const elem of document.getElementsByClassName("user-popup")) {
        elem.addEventListener("mousedown", function (event) {
            if(event.target.classList.contains("user-popup")) {
                closePopup()
            }}
        );
    }
}

async function checkHash() {
    if (location.hash === "#generate-ram" && user.username === userInfoUsername) {
        ramGenerator = new Generator();
    }
    if (location.hash === "#ram") {
        const url = new URL(location);
        ramPage = new RamPage(parseInt(url.searchParams.get("ram-id")))
    }
}

async function closeCanvas() {
    stage.destroy();
    stage = null;
    isCanvasMode = false;
    document.querySelector("#ram .popup-menu").innerHTML = ``
}

function closePopup() {
    try {
        if (ramGenerator) {
            ramGenerator.close();
        }
        if (ramPage) {
            ramPage.close();
        }
        if (isCanvasMode) {
            closeCanvas()
        }
    } catch (e) {}

    location.hash = "";

    const url = new URL(location);
    if (url.searchParams.has("ram-id")) {
        url.searchParams.delete("ram-id");
        history.pushState({}, "", url);
    }
}

window.addEventListener('beforeunload', closePopup);
