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