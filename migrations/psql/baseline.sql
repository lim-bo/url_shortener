CREATE TABLE "redirects" (
    id serial primary key,
    link TEXT NOT NULL,
    short_code VARCHAR(8) NOT NULL
);