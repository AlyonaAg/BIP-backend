# BIP-backend

## Запуск:
sudo docker-compose up —build bip_backend

## Требования к данным:
  1) *username*: обязательное поле, длина [2, 30], (a-zA-Z0-9);
  2) *password*: обязательное поле, длина [5, 100];
  3) *first_name*: обязательное поле, длина [1, 15], (a-zA-Z);
  4) *second_name*: обязательное поле, длина [1, 15], (a-zA-Z);
  5) *avatar_url*: необязательно поле, должен быть URL'ом;
  6) *phone_number*: необязательно поле, должен соответвовать E164 (+ и (0-9));
  7) *mail*: обязательно поле, должен быть почтой (@lala.lala обязательно)

## Для подключения к pgAdmin:
1) localhost:5050:
- Login: *pgadmin@pgadmin.org*
- Password: *admin*
2) Create -> Server
3) General:
- Name: bip_db
4) Connection:
- Host: db
- Port: 5432
- Username: postgres
- Password: admin

## Время жизни токенов/паролей
1) Токены актуальны в течение 24 часов
2) Одноразовые пароли для двухфакторной аутентификации актуальны 5 минут
