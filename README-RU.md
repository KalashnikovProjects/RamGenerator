# [<img src="images/icon512.png" width="40"/>](https://taprams.ru) RamGenerator
### Сайт для генерации и тапания баранов. https://taprams.ru

#### [English](README.md) Русский

<img src="images/index.png" width="600" alt="main page screenshot"/>

<details><summary>Скриншоты</summary>

<img src="images/top.png" width=600 alt="top rams section screenshot"/>
<img src="images/ram.png" width=600 alt="ram page screenshot"/>
<img src="images/generate-ram.png" width="600" alt="generate ram page screenshot"/>

</details>

## API Документация
### Swagger-ui - https://taprams.ru/swagger
[<img src="images/swagger.png" width="600"/>](https://taprams.ru/swagger)

## Стек технологий:
* Микросервисы - **<img src="images/Docker_Logo.png" height="13" alt="Docker"/> Docker**
* Обратный прокси **<img src="https://avatars.githubusercontent.com/u/12955528?s=48&v=4" height="13" alt="Caddy"/> Caddy**. Почему Caddy? Он слишком уж легко решает проблемы с SSL сертификатами.
* База данных - **<img src="https://upload.wikimedia.org/wikipedia/commons/thumb/2/29/Postgresql_elephant.svg/500px-Postgresql_elephant.svg.png" height="13" alt="Postgres"/> PostgreSQL**
* **<img src="https://camo.githubusercontent.com/99ff5593d0cf327a76225b5c99d4388c890936549e89b46ac019259c106f8c4f/68747470733a2f2f7261772e6769746875622e636f6d2f676f6c616e672d73616d706c65732f676f706865722d766563746f722f6d61737465722f676f706865722e706e67" height="13" alt="Go"/>  Go** - REST API сервер и сервер для статичных файлов
* **<img src="https://upload.wikimedia.org/wikipedia/commons/thumb/c/c3/Python-logo-notext.svg/3840px-Python-logo-notext.svg.png" height="13" alt="Postgres"/> Python** - gRPC сервер, делает запросы к API нейросетей (Gemini для текста, Cloudflare AI Workers для изображений)
* Frontend без фреймворков, простой js, html, css. Html файлы разделены на шаблоны: base, header, footer и сам контент 
страниц, динамический контент страниц отрисовываются на frontend'е с помощью js.
* [<img src="images/swagger_logo.png" height="13" alt="Swagger"/> Swagger ui](https://taprams.ru/swagger) - API документация

### На уровне микросервисов это работает так:
* **postgres** - база данных.
* **caddy** - обратный прокси.
* **go-api** - REST api сервер. По умолчанию на порту 8082. [DockerHub repository](https://hub.docker.com/repository/docker/kalashnik/ramgenerator-go-api)
* **go-static-server** - возвращает статичные файлы сайта, рендерит html шаблоны при старте. По умолчанию на порту 8081. [DockerHub](https://hub.docker.com/repository/docker/kalashnik/ramgenerator-go-static-server)
* **python-ai** - gRPC сервер для запросов к нейросетям на порту 50051. [DockerHub](https://hub.docker.com/repository/docker/kalashnik/ramgenerator-python-ai)
* **swagger** - swagger ui, на порту 8083. [DockerHub](https://hub.docker.com/repository/docker/kalashnik/ramgenerator-swagger)

[<img src="images/docker.png" width="450"/>](images/docker.png)

## ▶ Запуск локально
### `docker-compose up`
Если хотите запустить изменённую версию, или просто собрать свои изображения docker, используйте `docker-compose up --build`

### Необходимые переменные окружения (также смотрите [шаблон .env файла](template.env))
#### Для python-ai
`IMAGE_GENERATOR_API_KEY`, `GEMINI_API_KEY`, `GEMINI_API_KEY`, `GRPC_SECRET_TOKEN`, `GRPC_HOST`, `GRPC_PORT`
`GRPC_SECRET_TOKEN`, `GRPC_HOST`, `GRPC_PORT`

#### Для go-api
`IMAGE_CDN_API_KEY`, `IMAGE_CDN_OPEN_API`, `IMAGE_CDN_INTERNAL_API`, `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_HOST`,
`GRPC_SECRET_TOKEN`, `GRPC_HOST`, `GRPC_PORT`, `HMAC`

#### Для postgres
`POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_HOST`,

#### Для go-static-server
`API_URL`, `WEBSOCKET_PROTOCOL`

#### Для cdn
`IMAGE_CDN_API_KEY`

### Команды для генерации кода из [proto](proto/ram_generator.proto) файлов:
* `make go-grpc` - только для Go
* `make py-grpc` - только ля Python
* `make grpc` - для Go и Python
