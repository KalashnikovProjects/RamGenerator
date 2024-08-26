function timeToTapHref() {
    if (!user) {
        document.querySelector("#time-to-tap").href = "/login"
        return;
    }

    document.querySelector("#time-to-tap").href = `/users/${user.username}`
}