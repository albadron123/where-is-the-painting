-- only one author of a painting for now!
create table paintings (
id serial primary key,
title varchar(100) not null,
creation_year integer check(creation_year >= 0 and creation_year <= extract(year from current_date)),
where_to_find text not null,
picture_address varchar(100) not null,
author_id integer not null references authors(id),
museum_id integer not null references museums(id)
);

create table authors (
id serial primary key,
name varchar(100) not null,
birth_year integer check(birth_year >= 0 and birth_year <= extract(year from current_date)),
death_year integer check(birth_year >= 0 and birth_year <= extract(year from current_date)),
biography text
);

create table museums (
id serial primary key,
name varchar(100) not null unique,
web_page varchar(100),
verified boolean not null default false
);


create table users (
id serial primary key,
login varchar(100) unique,
password_hashed text
);

create table users_preferences (
user_id serial references users(id),
painting_id serial references paintings(id),
primary key (user_id, painting_id)
);

create table rights (
	user_id serial references users(id),
	museum_id serial references museums(id),
	give_rights boolean not null,
	change_paintings boolean not null,
	primary key (user_id, museum_id)
);


drop table paintings;

insert into paintings(title,creation_year,where_to_find,picture_address) 
values('Звездная ночь',1889,'Музей современного искусства, Нью-Йорк, постоянная экспозиция?','pictcha.png'),
	  ('Черный супрематический квадрат',1915,'Третьяковская галерея, Москва, экспозиция XX век','pictcha1.png');

insert into authors(name,birth_year,death_year,biography) 
values('Винсент ван Гог',1853, 1890, 'нидерландский живописец и график, одна из трёх главных фигур постимпрессионизма (наряду с Полем Сезанном и Полем Гогеном), чьё творчество оказало значительное влияние на живопись XX века.'),
	  ('Казимир Малевич',1879, null, 'русский и советский художник-авангардист польского происхождения, педагог, теоретик искусства, философ. Основоположник супрематизма — одного из крупнейших направлений абстракционизма.');

insert into museums(name, web_page) values
('Третьяковская галерея', 'www.tretyakovgallery.ru'),
('Museum of Modern Art', 'www.moma.org');

insert into painting_author_link values (1,3), (1,1), (2,2);

insert into painting_museum_link values (1,2), (2,1);

--выбор всех картин из определенного музея--
select * from paintings where id in (select painting_id from painting_museum_link where museum_id = 2);

--выбор картины по id--
select * from paintings where paintings.id = 1;
--выбор имен авторов картины по id--
select name from authors where id in (select author_id from painting_author_link where painting_id = 2);

select * from paintings;
select * from authors;
select * from museums;

select * from painting_author_link;


select * from authors where name like 'К%';

