CREATE TABLE "user"
(
    "id" serial NOT NULL PRIMARY KEY,
    "username" varchar(30)  UNIQUE NOT NULL,
    "password" text NOT NULL,
    "first_name" varchar(15) NOT NULL,
    "second_name" varchar(15) NOT NULL,
    "is_photographer" bool NOT NULL,
    "money" int NOT NULL CHECK ("money" >= 0) DEFAULT 0,
    "avatar_url" text,
    "phone_number" varchar(15),
    "mail" text,
    "rating" real CHECK ("rating" >= 0 AND "rating" <= 5)
);

CREATE TABLE "photo_url"
(
    "id" serial NOT NULL PRIMARY KEY,
    "user_id" int NOT NULL REFERENCES "user"("id"),
    "url" text NOT NULL
);

CREATE TABLE "comments"
(
    "id" serial NOT NULL PRIMARY KEY,
    "user_id" int NOT NULL REFERENCES "user"("id"),
    "user_com_id" int NOT NULL REFERENCES "user"("id"),
    "content" text NOT NULL
);

CREATE TABLE "order"
(
    "id" serial NOT NULL PRIMARY KEY,
    "client_id" int NOT NULL REFERENCES "user"("id"),
    "photographer_id" int NOT NULL REFERENCES "user"("id"),
    "order_cost" int NOT NULL CHECK ("order_cost" >= 0),
    "location" point NOT NULL,
    "client_current_location" point NOT NULL,
    "order_state" int NOT NULL
);
