function timeToTapHref() {
    if(loadingUser) {
        setTimeout(displayUser, 5);
        return;
    }

    if (!user) {
        document.querySelector("#time-to-tap").href = "/login"
        return;
    }

    document.querySelector("#time-to-tap").href = `/users/${user.username}`

}