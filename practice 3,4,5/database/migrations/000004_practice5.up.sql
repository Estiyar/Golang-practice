ALTER TABLE users
ADD COLUMN IF NOT EXISTS email VARCHAR(255),
ADD COLUMN IF NOT EXISTS age INT DEFAULT 0,
ADD COLUMN IF NOT EXISTS gender VARCHAR(20),
ADD COLUMN IF NOT EXISTS birth_date TIMESTAMP;

CREATE TABLE IF NOT EXISTS user_friends (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    friend_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, friend_id),
    CONSTRAINT no_self_friend CHECK (user_id <> friend_id)
);

INSERT INTO users (name, email, age, gender, birth_date)
VALUES
('Alice', 'alice@mail.com', 20, 'female', '2004-01-10'),
('Bob', 'bob@mail.com', 21, 'male', '2003-02-11'),
('Charlie', 'charlie@mail.com', 22, 'male', '2002-03-12'),
('Diana', 'diana@mail.com', 20, 'female', '2004-04-13'),
('Eve', 'eve@mail.com', 23, 'female', '2001-05-14'),
('Frank', 'frank@mail.com', 24, 'male', '2000-06-15'),
('Grace', 'grace@mail.com', 21, 'female', '2003-07-16'),
('Henry', 'henry@mail.com', 22, 'male', '2002-08-17'),
('Ivy', 'ivy@mail.com', 20, 'female', '2004-09-18'),
('Jack', 'jack@mail.com', 25, 'male', '1999-10-19'),
('Kate', 'kate@mail.com', 21, 'female', '2003-11-20'),
('Leo', 'leo@mail.com', 22, 'male', '2002-12-21'),
('Mia', 'mia@mail.com', 20, 'female', '2004-01-22'),
('Noah', 'noah@mail.com', 23, 'male', '2001-02-23'),
('Olivia', 'olivia@mail.com', 24, 'female', '2000-03-24'),
('Paul', 'paul@mail.com', 25, 'male', '1999-04-25'),
('Queen', 'queen@mail.com', 21, 'female', '2003-05-26'),
('Ryan', 'ryan@mail.com', 22, 'male', '2002-06-27'),
('Sara', 'sara@mail.com', 20, 'female', '2004-07-28'),
('Tom', 'tom@mail.com', 23, 'male', '2001-08-29')
ON CONFLICT DO NOTHING;

INSERT INTO user_friends (user_id, friend_id) VALUES
(1,3),(1,4),(1,5),(1,6),
(2,3),(2,4),(2,5),(2,7),
(3,1),(3,2),
(4,1),(4,2),
(5,1),(5,2),
(6,1),
(7,2)
ON CONFLICT DO NOTHING;