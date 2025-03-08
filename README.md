**Опиание:**

Веб-сервер, который реализует функциональность простейшего планировщика задач, аналог TODO-листа.

**Инструкция по запуску локально**
1. Установите пароль через переменную окружения:
   Для Windows (cmd):
   `set TODO_PASSWORD=12345`
   Для Linux/MacOS:
   `export TODO_PASSWORD=12345`
2. Запустите приложение:
   `go run cmd/main.go`
3. Откройте браузер по адресу:
   `http://localhost:7540`

**Запуск тестов**
1. Получите JWT-токен, отправив запрос (пароль меняем на свой):
   `curl -X POST http://localhost:7540/api/signin -H "Content-Type: application/json" -d "{\"password\": \"12345\"}"`
2. Вставьте полученный токен в файл config/settings.go:
   `var Token = ваш_токен`
3. Запустите тесты:
   `go test ./tests`

**Docker-сборка**
`docker build -t todo_scheduler .`

**Запуск контейнера Docker**
1. База данных в проекте находится: `D:\Dev\go_final_project\data`
2. Запустите контейнер следующей командой:
   `docker run -d -p 7540:7540 -v D:\Dev\go_final_project\data:/app/data -e TODO_PASSWORD=12345 todo_scheduler`
2. После этого приложение будет доступно по адресу:
   `http://localhost:7540`