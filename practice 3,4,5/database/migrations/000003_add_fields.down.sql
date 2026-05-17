alter table users
drop column if exists email,
drop column if exists age,
drop column if exists created_at;