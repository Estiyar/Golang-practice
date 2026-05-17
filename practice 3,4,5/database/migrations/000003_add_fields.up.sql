alter table users
add column if not exists email varchar(255) not null default '',
add column if not exists age int not null default 0,
add column if not exists created_at timestamptz not null default now();