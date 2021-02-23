create extension if not exists citext;
create extension if not exists "uuid-ossp";

drop table if exists users cascade;
drop table if exists lessons cascade;
drop table if exists days cascade;
drop table if exists weeks cascade;
drop table if exists groups cascade;
drop table if exists students cascade;
drop table if exists timetables cascade;

drop type if exists users_roles;
drop type if exists lessons_types;
drop type if exists week_types;


/* guest role - for testing */
create type users_roles as enum ('admin', 'methodist', 'student', 'professor', 'guest');
create type lessons_types as enum ('sem', 'lek', 'lr', 'dz', 'rk', 'cons', 'exam', 'free');
create type week_types as enum ('numerator', 'denominator'); /* Числитель и знаменатель */

create unlogged table users
(
    user_id        uuid        default uuid_generate_v4()
        constraint user_id_pkey primary key not null,
    role           users_roles default 'guest',
    email          citext collate "C"       not null unique,
    password_hash  bytea                    not null,
    name           varchar(128)             not null,
    surname        varchar(128)             not null,
    patronymic     varchar(128)             not null,
    phone          varchar(18),
    social_network text,
    about          text
);

create unlogged table lessons
(
    lesson_id   uuid         default uuid_generate_v4()
        constraint lesson_id_pkey primary key not null,
    title       citext collate "C"            not null,
    auditorium  varchar(10)  default null,
    lesson_type lessons_types                 not null,
    comment     varchar(255) default null
);

create unlogged table days
(
    day_id     uuid default uuid_generate_v4()
        constraint day_id_pkey primary key not null,
    lesson1_id uuid                        references lessons (lesson_id) on delete set null,
    lesson2_id uuid                        references lessons (lesson_id) on delete set null,
    lesson3_id uuid                        references lessons (lesson_id) on delete set null,
    lesson4_id uuid                        references lessons (lesson_id) on delete set null,
    lesson5_id uuid                        references lessons (lesson_id) on delete set null,
    lesson6_id uuid                        references lessons (lesson_id) on delete set null,
    lesson7_id uuid                        references lessons (lesson_id) on delete set null,
    lesson8_id uuid                        references lessons (lesson_id) on delete set null
);

create unlogged table weeks
(
    week_id   uuid       default uuid_generate_v4()
        constraint week_id_pkey primary key not null,
    week_type week_types default 'numerator',

    day1_id   uuid                          references days (day_id) on delete set null,
    day2_id   uuid                          references days (day_id) on delete set null,
    day3_id   uuid                          references days (day_id) on delete set null,
    day4_id   uuid                          references days (day_id) on delete set null,
    day5_id   uuid                          references days (day_id) on delete set null,
    day6_id   uuid                          references days (day_id) on delete set null,
    day7_id   uuid                          references days (day_id) on delete set null
);

create unlogged table groups
(
    group_id     uuid       default uuid_generate_v4()
        constraint group_id_pkey primary key not null,
    group_name   citext collate "C",
    timetable_id uuid       default null,
    year         varchar(5) default date_part('year', CURRENT_DATE),
    unique (group_name, year)
);

create unlogged table timetables
(
    timetable_id uuid default uuid_generate_v4()
        constraint timetable_id_pkey primary key   not null,
    group_id     uuid references groups (group_id) not null,

    week1_id     uuid                              references weeks (week_id) on delete set null,
    week2_id     uuid                              references weeks (week_id) on delete set null
--     week3_id     uuid references weeks (week_id),
--     week4_id     uuid references weeks (week_id),
--     week5_id     uuid references weeks (week_id),
--     week6_id     uuid references weeks (week_id),
--     week7_id     uuid references weeks (week_id),
--     week8_id     uuid references weeks (week_id),
--     week9_id     uuid references weeks (week_id),
--     week10_id    uuid references weeks (week_id),
--     week11_id    uuid references weeks (week_id),
--     week12_id    uuid references weeks (week_id),
--     week13_id    uuid references weeks (week_id),
--     week14_id    uuid references weeks (week_id),
--     week15_id    uuid references weeks (week_id),
--     week16_id    uuid references weeks (week_id),
--     week17_id    uuid references weeks (week_id)
);

create unlogged table students
(
    student_id uuid default uuid_generate_v4()
        constraint student_id_pkey primary key                     not null,
    group_id   uuid references groups (group_id) on delete cascade not null,
    user_id    uuid references users (user_id) on delete cascade   not null
);

