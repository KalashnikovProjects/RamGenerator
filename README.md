# [<img src="images/icon512.png" width="40"/>](https://taprams.ru) RamGenerator
### A site for generating and tapping rams. https://taprams.ru

#### English [Русский](README-RU.md )

<img src="images/index.png" width="600" alt="main page screenshot"/>

<details><summary>Screenshots</summary>

<img src="images/top.png" width=600 alt="top rams section screenshot"/>
<img src="images/ram.png" width=600 alt="ram page screenshot"/>
<img src="images/generate-ram.png" width="600" alt="generate ram page screenshot"/>

</details>

## API Documentation
### Swagger-ui - https://taprams.ru/swagger
[<img src="images/swagger.png" width="600"/>](https://taprams.ru/swagger)


## Technology stack:
* Microservices - **<img src="images/Docker_Logo.png" height="13" alt="Docker"/> Docker**
* Reverse proxy **<img src="https://avatars.githubusercontent.com/u/12955528?s=48&v=4 " height="13" alt="Caddy"/>
Caddy**. Why Caddy? It easily solves problems with SSL certificates.
* Database - **<img src="https://upload.wikimedia.org/wikipedia/commons/thumb/2/29/Postgresql_elephant.svg/500px-Postgresql_elephant.svg.png " height="13" alt="Postgres"/> 
PostgreSQL**
* **<img src="https://camo.githubusercontent.com/99ff5593d0cf327a76225b5c99d4388c890936549e89b46ac019259c106f8c4f/68747470733a2f2f7261772e6769746875622e636f6d2f676f6c616e672d73616d706c65732f676f706865722d766563746f722f6d61737465722f676f706865722e706e67 " height="13" alt="Go"/>
Go** - REST API server and server for static files
* **<img src="https://upload.wikimedia.org/wikipedia/commons/thumb/c/c3/Python-logo-notext.svg/3840px-Python-logo-notext.svg.png " height="13" alt="Postgres"/>
Python** is a gRPC server that makes requests to the neural network API (Gemini for text, Cloudflare AI Workers for images)
* Frontend without frameworks, simple js, html, css. Html files are divided into templates: base, header, footer and
  the page content itself, dynamic page content is rendered on the frontend using js.
* [<img src="images/swagger_logo.png" height="13" alt="Swagger"/> Swagger ui](https://taprams.ru/swagger ) - API documentation


### At the microservices/docker containers level, it works like this:
* **postgres** - database.
* **caddy** - reverse proxy.
* **go-api** - REST api server. By default on port 8082. [DockerHub repository](https://hub.docker.com/repository/docker/kalashnik/ramgenerator-go-api)
* **go-static-server** - returns static site files, renders html templates at startup. By default on port 8081. [DockerHub](https://hub.docker.com/repository/docker/kalashnik/ramgenerator-go-static-server)
* **python-ai** - gRPC server for requests to ai api's on port 50051. [DockerHub](https://hub.docker.com/repository/docker/kalashnik/ramgenerator-python-ai)
* **swagger** - swagger ui, use port 8083. [DockerHub](https://hub.docker.com/repository/docker/kalashnik/ramgenerator-swagger)

[<img src="images/docker.png" width="450"/>](images/docker.png)

## ▶ Local launch
### `docker-compose up`
If you want to run a modified version, or just build your docker images, use `docker-compose up --build`

### Necessary environment variables (see alse [template.env](template.env))
#### For python-ai
`IMAGE_GENERATOR_API_KEY`, `GEMINI_API_KEY`, `GEMINI_API_KEY`, `GRPC_SECRET_TOKEN`, `GRPC_HOST`, `GRPC_PORT`
`GRPC_SECRET_TOKEN`, `GRPC_HOST`, `GRPC_PORT`

#### For go-api
`IMAGE_CDN_API_KEY`, `IMAGE_CDN_OPEN_API`, `IMAGE_CDN_INTERNAL_API`, `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_HOST`,
`GRPC_SECRET_TOKEN`, `GRPC_HOST`, `GRPC_PORT`, `HMAC`

#### For postgres
`POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_HOST`,

#### For go-static-server
`API_URL`, `WEBSOCKET_PROTOCOL`

#### For cdn
`IMAGE_CDN_API_KEY`

### Generate code from [proto](proto/ram_generator.proto) files:
* `make go-grpc` - only for Go
* `make py-grpc` - only for Python
* `make grpc` - for Go and Python
