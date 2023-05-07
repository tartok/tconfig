create sequence users_id_seq as integer;

CREATE TABLE config
(
    "version" varchar(25) NOT NULL,
    "locale"  varchar(25) NOT NULL,
    PRIMARY KEY (version)
);
insert into config(version, locale)
values ('0.0.0.0', 'ua');

create table users
(
    id   int4         not null DEFAULT nextval('users_id_seq'::regclass),
    login varchar(25) not null ,
    password varchar(250),
    primary key (id),
    unique (login)
)