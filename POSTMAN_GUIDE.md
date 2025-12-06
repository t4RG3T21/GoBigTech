# Postman Guide - Тестирование GoBigTech через Envoy

Краткое руководство по тестированию всех API через Postman.

## 🚀 Быстрый старт

1. **Запустите все сервисы:**
   ```powershell
   docker-compose up -d
   ```

2. **Дождитесь готовности:**
   ```powershell
   docker ps
   # Все сервисы должны быть (healthy)
   ```

3. **Откройте Postman** и импортируйте коллекцию (см. ниже)

## 📡 Базовый URL

Все запросы идут через **Envoy API Gateway**:

```
http://localhost
```

**НЕ используйте** прямые порты сервисов (8080, 8082, etc.) - все через Envoy на порту 80!

## 🔐 Аутентификация

Большинство эндпоинтов требуют токен в заголовке `Authorization`:

```
Authorization: Bearer YOUR_SESSION_TOKEN
```

## 📋 Коллекция запросов

### 1. Health Check (публичный)

**GET** `http://localhost/health`

**Headers:** (не требуются)

**Response:**
```json
{
  "status": "ok",
  "service": "order",
  "timestamp": "2025-12-05T10:00:00Z"
}
```

---

### 2. Регистрация пользователя (публичный)

**POST** `http://localhost/api/iam/v1/iam/register`

**Headers:**
```
Content-Type: application/json
```

**Body (JSON):**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "username": "testuser"
}
```

**Response:**
```json
{
  "success": true,
  "user_id": "user-123",
  "message": "User registered successfully"
}
```

**Сохраните `user_id` для дальнейших запросов!**

---

### 3. Вход пользователя (публичный)

**POST** `http://localhost/api/iam/v1/iam/login`

**Headers:**
```
Content-Type: application/json
```

**Body (JSON):**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Или по username:**
```json
{
  "username": "testuser",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "session_id": "abc123def456...",
  "user_id": "user-123",
  "message": "Login successful"
}
```

**⚠️ ВАЖНО: Сохраните `session_id` - это ваш токен для аутентификации!**

---

### 4. Валидация сессии

**POST** `http://localhost/api/iam/v1/iam/validate-session`

**Headers:**
```
Content-Type: application/json
Authorization: Bearer YOUR_SESSION_TOKEN
```

**Body (JSON):**
```json
{
  "session_id": "YOUR_SESSION_TOKEN"
}
```

**Response:**
```json
{
  "valid": true,
  "user_id": "user-123",
  "email": "user@example.com"
}
```

---

### 5. Получение информации о пользователе

**GET** `http://localhost/api/iam/v1/iam/user/{user_id}`

**Headers:**
```
Authorization: Bearer YOUR_SESSION_TOKEN
```

**Path Variables:**
- `user_id` - ID пользователя (из ответа регистрации/логина)

**Response:**
```json
{
  "user_id": "user-123",
  "email": "user@example.com",
  "name": "testuser",
  "created_at": 1234567890
}
```

---

### 6. Создание заказа (требует аутентификацию)

**POST** `http://localhost/api/orders`

**Headers:**
```
Content-Type: application/json
Authorization: Bearer YOUR_SESSION_TOKEN
```

**Body (JSON):**
```json
{
  "user_id": "user-123",
  "items": [
    {
      "product_id": "prod-123",
      "quantity": 2
    },
    {
      "product_id": "prod-456",
      "quantity": 1
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "id": "order-123",
  "user_id": "user-123",
  "status": "pending",
  "items": [
    {
      "product_id": "prod-123",
      "quantity": 2
    }
  ]
}
```

**⚠️ Если получили 401 Unauthorized:**
- Проверьте, что токен правильный
- Убедитесь, что используете формат `Bearer TOKEN`
- Проверьте, что токен не истек

---

### 7. Получение заказа по ID

**GET** `http://localhost/api/orders/{id}`

**Headers:**
```
Authorization: Bearer YOUR_SESSION_TOKEN
```

**Path Variables:**
- `id` - ID заказа (из ответа создания заказа)

**Response:**
```json
{
  "id": "order-123",
  "user_id": "user-123",
  "status": "completed",
  "items": [...]
}
```

---

## 🔧 Настройка Postman Environment

Создайте Environment в Postman для удобства:

### Variables:

| Variable | Initial Value | Current Value |
|----------|---------------|---------------|
| `base_url` | `http://localhost` | `http://localhost` |
| `session_token` | (пусто) | (будет заполнено после логина) |
| `user_id` | (пусто) | (будет заполнено после регистрации) |

### Использование:

В URL используйте: `{{base_url}}/api/orders`

В Headers используйте: `Authorization: Bearer {{session_token}}`

### Автоматическое сохранение токена:

1. Создайте тест для запроса Login:
   ```javascript
   if (pm.response.code === 200) {
       const jsonData = pm.response.json();
       pm.environment.set("session_token", jsonData.session_id);
       pm.environment.set("user_id", jsonData.user_id);
   }
   ```

2. Теперь токен будет автоматически использоваться в следующих запросах!

---

## 🧪 Тестовые сценарии

### Сценарий 1: Полный цикл создания заказа

1. ✅ **Регистрация** → получите `user_id`
2. ✅ **Логин** → получите `session_token`
3. ✅ **Создание заказа** → используйте `session_token` в заголовке
4. ✅ **Получение заказа** → проверьте статус

### Сценарий 2: Проверка защиты эндпоинтов

1. ❌ **Создание заказа БЕЗ токена** → должен вернуть `401 Unauthorized`
2. ❌ **Создание заказа с невалидным токеном** → должен вернуть `401 Unauthorized`
3. ✅ **Создание заказа с валидным токеном** → должен вернуть `200 OK`

### Сценарий 3: Публичные эндпоинты

1. ✅ **Health check** → должен работать без токена
2. ✅ **Регистрация** → должен работать без токена
3. ✅ **Логин** → должен работать без токена

---

## 🐛 Troubleshooting

### Ошибка 401 Unauthorized

**Проблема:** Запрос возвращает 401

**Решение:**
1. Проверьте заголовок `Authorization: Bearer TOKEN`
2. Убедитесь, что токен не истек (сессия живет 24 часа)
3. Проверьте логи Envoy: `docker logs gobigtech_envoy --tail 50`
4. Проверьте логи IAM: `docker logs gobigtech_iam_service --tail 50`

### Ошибка 404 Not Found

**Проблема:** Запрос возвращает 404

**Решение:**
1. Проверьте URL - должен быть `/api/orders`, а не `/orders`
2. Проверьте, что Envoy запущен: `docker ps | grep envoy`
3. Проверьте маршрутизацию в Envoy: `docker logs gobigtech_envoy --tail 50`

### Ошибка 500 Internal Server Error

**Проблема:** Запрос возвращает 500

**Решение:**
1. Проверьте логи сервиса: `docker logs gobigtech_order_service --tail 50`
2. Проверьте, что все зависимости запущены (PostgreSQL, Kafka, etc.)
3. Проверьте health checks: `docker ps`

### Ошибка Connection Refused

**Проблема:** Не удается подключиться

**Решение:**
1. Проверьте, что Docker запущен
2. Проверьте, что все сервисы запущены: `docker-compose ps`
3. Проверьте порт 80: `netstat -an | findstr :80`

---

## 📊 Проверка через Postman Console

Включите Postman Console (View → Show Postman Console) для просмотра:
- Полных заголовков запросов
- Тела запросов
- Ответов серверов
- Времени выполнения

---

## 🎯 Импорт коллекции

Создайте новую коллекцию в Postman и добавьте все запросы выше. Или используйте этот JSON:

```json
{
  "info": {
    "name": "GoBigTech Envoy API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost"
    },
    {
      "key": "session_token",
      "value": ""
    }
  ],
  "item": [
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{base_url}}/health",
          "host": ["{{base_url}}"],
          "path": ["health"]
        }
      }
    },
    {
      "name": "Register",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"email\": \"user@example.com\",\n  \"password\": \"password123\",\n  \"username\": \"testuser\"\n}"
        },
        "url": {
          "raw": "{{base_url}}/api/iam/v1/iam/register",
          "host": ["{{base_url}}"],
          "path": ["api", "iam", "v1", "iam", "register"]
        }
      }
    },
    {
      "name": "Login",
      "event": [
        {
          "listen": "test",
          "script": {
            "exec": [
              "if (pm.response.code === 200) {",
              "    const jsonData = pm.response.json();",
              "    pm.environment.set(\"session_token\", jsonData.session_id);",
              "    pm.environment.set(\"user_id\", jsonData.user_id);",
              "}"
            ]
          }
        }
      ],
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"email\": \"user@example.com\",\n  \"password\": \"password123\"\n}"
        },
        "url": {
          "raw": "{{base_url}}/api/iam/v1/iam/login",
          "host": ["{{base_url}}"],
          "path": ["api", "iam", "v1", "iam", "login"]
        }
      }
    },
    {
      "name": "Create Order",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          },
          {
            "key": "Authorization",
            "value": "Bearer {{session_token}}"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"user_id\": \"{{user_id}}\",\n  \"items\": [\n    {\n      \"product_id\": \"prod-123\",\n      \"quantity\": 2\n    }\n  ]\n}"
        },
        "url": {
          "raw": "{{base_url}}/api/orders",
          "host": ["{{base_url}}"],
          "path": ["api", "orders"]
        }
      }
    }
  ]
}
```

Сохраните это в файл `GoBigTech.postman_collection.json` и импортируйте в Postman!

---

## ✅ Чеклист тестирования

- [ ] Health check работает
- [ ] Регистрация пользователя работает
- [ ] Логин возвращает токен
- [ ] Создание заказа с токеном работает
- [ ] Создание заказа без токена возвращает 401
- [ ] Получение заказа работает
- [ ] Все запросы идут через Envoy (порт 80)

**Готово! Теперь вы можете тестировать весь API через Postman!** 🎉

