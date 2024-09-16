let stage, layer, image, tr, cropLayer, overlayLayer;
let isCanvasMode = false;
const strokeWidth = 2;
const overlayColor = 'rgba(21,8,33,0.71)';

function toggleCropMode() {
    isCanvasMode = !isCanvasMode;
    if (isCanvasMode) {
        toCanvas();
    } else {
        toImg();
    }
}

function toImg() {
    if (stage) {
        stage.destroy();
        stage = null;
    }
    const url = new URL(location);
    ramPage = new RamPage(parseInt(url.searchParams.get("ram-id")));
}

function toCanvas() {
    let img = document.getElementById('ram-clicker');
    const width = img.width;
    const height = img.height;
    ramPage.close(false);
    document.getElementById("ram-content").innerHTML = `<div id="ram-crop" class="ram-image mt-5"></div>
            <button id="save-avatar" class="button-user" onclick="saveAvatar()">Сохранить аватар</button>`;

    stage = new Konva.Stage({
        container: 'ram-crop',
        width: width,
        height: height,
    });

    layer = new Konva.Layer();
    stage.add(layer);

    overlayLayer = new Konva.Layer();
    stage.add(overlayLayer);

    cropLayer = new Konva.Layer();
    stage.add(cropLayer);

    const imageObj = new Image();
    imageObj.src = img.src;
    imageObj.onload = function() {
        let imageNode = new Konva.Image({
            image: imageObj,
            x: 0,
            y: 0,
            width: stage.width(),
            height: stage.height()
        });

        layer.add(imageNode);

        const padding = 50;
        const size = Math.min(stage.width(), stage.height()) - padding * 2;
        let cropRect = new Konva.Rect({
            x: padding,
            y: padding,
            width: size,
            height: size,
            stroke: '#2e0055',
            shouldOverdrawWholeArea: true,
            strokeWidth: strokeWidth,
            draggable: true
        });

        cropLayer.add(cropRect);

        tr = new Konva.Transformer({
            borderStroke: `#2e0055`,
            nodes: [cropRect],
            keepRatio: true,
            enabledAnchors: ['top-left', 'top-right', 'bottom-left', 'bottom-right'],
            borderEnabled: false,
            rotateEnabled: false,
            anchorFill: '#8000ff', // Фиолетовый цвет для углов
            anchorStroke: '#2e0055',
            boundBoxFunc: function (oldBox, newBox) {
                if (
                    newBox.x < -strokeWidth ||
                    newBox.y < -strokeWidth ||
                    newBox.x + newBox.width > stage.width() + strokeWidth ||
                    newBox.y + newBox.height > stage.height() + strokeWidth ||
                    newBox.width <= 0 ||
                    newBox.height <= 0
                ) {
                    return oldBox;
                }
                updateMask()
                return newBox;
            }
        });
        cropLayer.add(tr);

        cropRect.on('dragmove', function() {
            const pos = this.position();
            this.position({
                x: Math.max(0, Math.min(pos.x, stage.width() - this.width() * this.scaleX())),
                y: Math.max(0, Math.min(pos.y, stage.height() - this.height() * this.scaleY()))
            });
            updateMask()
        });
        cropRect.on('transform', updateMask);

        cropRect.on('mouseover', function () {
            document.body.style.cursor = 'move';
        });
        cropRect.on('mouseout', function () {
            document.body.style.cursor = 'default';
        });

        updateMask();
        layer.draw();
        cropLayer.draw();
    };
}

function updateMask() {
    overlayLayer.destroyChildren();

    const rects = [
        new Konva.Rect({
            x: 0,
            y: 0,
            width: stage.width(),
            height: tr.y(),
            fill: overlayColor
        }),
        new Konva.Rect({
            x: 0,
            y: tr.y(),
            width: tr.x(),
            height: tr.height(),
            fill: overlayColor
        }),
        new Konva.Rect({
            x: tr.x() + tr.width(),
            y: tr.y(),
            width: stage.width() - (tr.x() + tr.width()),
            height: tr.height(),
            fill: overlayColor
        }),
        new Konva.Rect({
            x: 0,
            y: tr.y() + tr.height(),
            width: stage.width(),
            height: stage.height() - (tr.y() + tr.height()),
            fill: overlayColor
        }),
    ];

    rects.forEach(rect => overlayLayer.add(rect));

    overlayLayer.draw();
}

function saveAvatar() {
    let flatCoords = [
        tr.x() / stage.width(), tr.y() / stage.height(),
        (tr.x() + tr.width()) / stage.width(), (tr.y() + tr.height()) / stage.height()]
    flatCoords.forEach((val, ind, array) => {
        if (val < 0) {
            array[ind] = 0
        } else if (val > 1) {
            array[ind] = 1
        }
    })
    let coords = [
        [flatCoords[0], flatCoords[1]],
        [flatCoords[2], flatCoords[3]]]
    const el = document.getElementById('username-message')

    fetch(`${API_URL}/users/${userInfoUsername}`, {
        mode: 'cors',
        method: 'PATCH',
        headers: {
            'Authorization': `Bearer ${getCookie("token")}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({"avatar_ram_id": parseInt((new URL(location)).searchParams.get("ram-id")), "avatar_box": coords})
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