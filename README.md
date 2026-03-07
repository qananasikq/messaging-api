# 📱 **Messaging API**

Backend-сервис для обмена сообщениями с поддержкой реального времени, использующий WebSocket для отправки и получения сообщений.

## 🚀 Основные возможности

* **Регистрация и аутентификация пользователей** с использованием JWT
* **Создание личных и групповых чатов**
* **Реал-тайм обновления** через WebSocket
* **Счётчики непрочитанных сообщений**
* **Удаление диалога** пользователем-создателем
* Полностью **контейнеризированное окружение** (Docker Compose)

---

##  Стек технологий

* **Backend**: Go, Gin, PostgreSQL, Redis, WebSocket, JWT
* **Frontend**: React, Vite
* **Инфраструктура**: Docker, Docker Compose

---

## 📂 Структура проекта

```plaintext
messaging-api/
├── cmd/
│   ├── api/          # Точка входа сервера
│   └── migrate/      # Запуск миграций базы данных
├── internal/
│   ├── config/       # Конфигурация приложения
│   ├── handlers/     # HTTP и WebSocket обработчики
│   ├── middleware/   # Middleware для API
│   ├── repositories/ # Доступ к данным (PostgreSQL)
│   ├── services/     # Логика работы приложения
│   └── websocket/    # WebSocket-хаб и клиенты
├── migrations/       # Миграции базы данных
├── pkg/              # Утилиты, например, для работы с JWT
├── web/              # Frontend (React)
├── .dockerignore     # Исключаемые файлы для Docker
├── .env              # Переменные окружения
├── .env.example      # Пример конфигурации для .env
├── .gitignore        # Исключаемые файлы для git
├── Dockerfile        # Инструкция по сборке Docker-образа
├── docker-compose.yml # Конфигурация для Docker Compose
├── go.mod            # Зависимости Go
├── go.sum            # Контроль зависимостей Go
├── Makefile          # Автоматизация задач
└── README.md         # Документация проекта
```

---

## 🛠 Установка

### 📋 Требования

* **Go 1.18+**
* **Docker**
* **Docker Compose**

### 🔧 Локальная установка

1. **Склонируйте репозиторий**:

   ```bash
   git clone https://github.com/yourusername/messaging-api.git
   ```

2. **Перейдите в каталог проекта**:

   ```bash
   cd messaging-api
   ```

3. **Создайте файл `.env`**, основываясь на `.env.example`, и настройте переменные окружения.

4. **Соберите и запустите приложение** с помощью Docker Compose:

   ```bash
   docker-compose up --build
   ```

5. После того как контейнеры запустятся, **API будет доступно по адресу** `http://localhost:8080`.

---

### 🛠 Миграции

Чтобы применить миграции в базе данных, используйте команду:

```bash
docker-compose exec api go run cmd/migrate/main.go
```

---

## 💻 Пример использования API

1. **Регистрация нового пользователя**:

   Запрос:

   ```bash
   POST /users
   {
       "username": "example",
       "password": "password123"
   }
   ```

   Ответ:

   ```json
   {
       "id": "uuid",
       "username": "example",
       "token": "jwt_token"
   }
   ```

2. **Получение информации о пользователе**:

   Запрос:

   ```bash
   GET /users/:id
   ```

   Ответ:

   ```json
   {
       "id": "uuid",
       "username": "example"
   }
   ```

---

## 🛠 Разработка

Для внесения изменений и разработки новых фич:

1. **Склонируйте репозиторий**.

2. **Создайте отдельную ветку** для работы:

   ```bash
   git checkout -b feature/your-feature-name
   ```

3. Убедитесь, что ваш код протестирован (планируется добавить тесты в будущих релизах).

4. **Создайте Pull Request** в основную ветку.

---

## 📝 Лицензия

MIT License.

---