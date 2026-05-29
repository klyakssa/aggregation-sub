# aggregation-sub
REST-сервис для агрегации данных об онлайн подписках пользователей

Для запуска используйте:
make docker-up


Пример запроса для создания подписки
curl -X POST http://localhost:8080/api/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Netflix",
    "price": 799,
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "start_date": "01-2024"
  }'


Для просмотра базы данных используется дополнительный контейнер с PgAdmin