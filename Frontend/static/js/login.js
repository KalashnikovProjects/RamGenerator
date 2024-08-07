const LOGIN_ERROR_ELEMENT_ID = 'loginError';
const REGISTER_ERROR_ELEMENT_ID = 'registerError';


function displayError(elementId, message) {
    const errorElement = document.getElementById(elementId);
    if (errorElement) {
        errorElement.textContent = message;
    } else {
        console.error('Error element not found:', elementId, message);
    }
}

function validateUsername(username) {
    return username.length >= 3 && username.length <= 24;
}

function validatePassword(password, passwordRepeat = null) {
    if (password === "") {
        return "Необходимо заполнить поле пароль";
    }
    if (passwordRepeat !== null && password !== passwordRepeat) {
        return "Пароли не совпадают";
    }
    return null;
}

async function handleServerResponse(response, successCallback, errorElementId) {
    const text = await response.text();
    if (response.ok) {
        setCookie("token", text, {samesite: "lax"});
        sessionStorage.removeItem("user")
        successCallback(await loadUser());
    } else {
        console.log(response, text);
        let errorText;

        switch (response.status) {
            case 404:
                errorText = "Такого пользователя не существует";
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
            case 401:
                errorText = "Неверный пароль";
                break;
            default:
                errorText = "Произошла неизвестная ошибка на сервере, попробуйте войти позже";
        }

        displayError(errorElementId, errorText);
    }
}

async function sendRequest(url, data, successCallback, errorElementId) {
    try {
        const response = await fetch(url, {
            mode: 'cors',
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data)
        });
        await handleServerResponse(response, successCallback, errorElementId);
    } catch (error) {
        console.error('Request error:', error);
        displayError(errorElementId, "Произошла ошибка при подключении к серверу");
    }
}

function handleLogin(event) {
    event.preventDefault();

    const username = document.getElementById('loginUsername').value;
    const password = document.getElementById('loginPassword').value;

    if (!validateUsername(username)) {
        displayError(LOGIN_ERROR_ELEMENT_ID, "Имя должно содержать от 3 до 24 символов");
        return;
    }

    const passwordError = validatePassword(password);
    if (passwordError) {
        displayError(LOGIN_ERROR_ELEMENT_ID, passwordError);
        return;
    }

    sendRequest(`${API_URL}/login`, { username, password },
        (user) => {
            window.location.href = `/users/${user.username}`;
        },
        LOGIN_ERROR_ELEMENT_ID
    );
}

function handleRegister(event) {
    event.preventDefault();

    const username = document.getElementById('registerUsername').value;
    const password = document.getElementById('registerPassword').value;
    const passwordRepeat = document.getElementById('registerPasswordRepeat').value;

    if (!validateUsername(username)) {
        displayError(REGISTER_ERROR_ELEMENT_ID, "Имя должно содержать от 3 до 24 символов");
        return;
    }

    const passwordError = validatePassword(password, passwordRepeat);
    if (passwordError) {
        displayError(REGISTER_ERROR_ELEMENT_ID, passwordError);
        return;
    }

    sendRequest(`${API_URL}/register`, { username, password },
        (user) => {
            window.location.href = new URLSearchParams(window.location.search).get('redirect') ?? `/users/${user.username}`;
        },
        REGISTER_ERROR_ELEMENT_ID
    );
}

document.addEventListener('DOMContentLoaded', () => {
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }

    const registerForm = document.getElementById('registerForm');
    if (registerForm) {
        registerForm.addEventListener('submit', handleRegister);
    }
});