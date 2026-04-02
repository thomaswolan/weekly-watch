-- ============================================================
-- The Weekly Watch: A Movie Recommendation System
-- Phase 5: Database Connectivity and User Interface Operations
-- CS 4604 - Spring 2026
-- DBMS: MySQL 8.x
-- ============================================================

DROP DATABASE IF EXISTS weekly_watch;
CREATE DATABASE weekly_watch;
USE weekly_watch;

-- ============================================================
-- TABLE DEFINITIONS (ordered by dependency)
-- ============================================================

CREATE TABLE User (
    user_id       INT           PRIMARY KEY AUTO_INCREMENT,
    username      VARCHAR(50)   NOT NULL UNIQUE,
    email         VARCHAR(100)  NOT NULL UNIQUE,
    created_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login    DATETIME      DEFAULT NULL
);

CREATE TABLE Movie (
    movie_id      INT           PRIMARY KEY AUTO_INCREMENT,
    title         VARCHAR(255)  NOT NULL,
    plot_summary  TEXT,
    trailer_url   VARCHAR(500),
    tmdb_id       VARCHAR(20)   UNIQUE
);

CREATE TABLE Person (
    person_id        INT           PRIMARY KEY AUTO_INCREMENT,
    full_name        VARCHAR(150)  NOT NULL,
    birth_date       DATE,
    biography        TEXT,
    profile_image_url VARCHAR(500),
    imdb_id          VARCHAR(20)   UNIQUE
);

CREATE TABLE Role (
    role_id       INT           PRIMARY KEY AUTO_INCREMENT,
    role_name     VARCHAR(50)   NOT NULL UNIQUE,
    description   TEXT
);

CREATE TABLE Genre (
    genre_id      INT           PRIMARY KEY AUTO_INCREMENT,
    genre_name    VARCHAR(50)   NOT NULL UNIQUE,
    description   TEXT
);

CREATE TABLE Streaming_Service (
    service_id    INT           PRIMARY KEY AUTO_INCREMENT,
    service_name  VARCHAR(100)  NOT NULL UNIQUE
);

CREATE TABLE Notification (
    notification_id   INT           PRIMARY KEY AUTO_INCREMENT,
    user_id           INT           NOT NULL,
    notification_type VARCHAR(50)   NOT NULL,
    message_text      TEXT          NOT NULL,
    sent_at           DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status            VARCHAR(20)   NOT NULL DEFAULT 'unread',
    FOREIGN KEY (user_id) REFERENCES User(user_id) ON DELETE CASCADE
);

CREATE TABLE User_Preference (
    preference_id    INT           PRIMARY KEY AUTO_INCREMENT,
    user_id          INT           NOT NULL,
    preference_type  VARCHAR(50)   NOT NULL,
    preference_value VARCHAR(100)  NOT NULL,
    strength         VARCHAR(20)   DEFAULT 'medium',
    created_at       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES User(user_id) ON DELETE CASCADE
);

CREATE TABLE Favorites_List (
    favorites_id  INT           PRIMARY KEY AUTO_INCREMENT,
    user_id       INT           NOT NULL,
    list_name     VARCHAR(100)  NOT NULL DEFAULT 'My Favorites',
    created_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES User(user_id) ON DELETE CASCADE
);

CREATE TABLE Review (
    review_id     INT           PRIMARY KEY AUTO_INCREMENT,
    user_id       INT           NOT NULL,
    movie_id      INT           NOT NULL,
    review_text   TEXT          NOT NULL,
    created_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    is_spoiler    BOOLEAN       NOT NULL DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES User(user_id) ON DELETE CASCADE,
    FOREIGN KEY (movie_id) REFERENCES Movie(movie_id) ON DELETE CASCADE
);

CREATE TABLE Rating (
    rating_id     INT           PRIMARY KEY AUTO_INCREMENT,
    user_id       INT           NOT NULL,
    movie_id      INT           NOT NULL,
    rating_value  VARCHAR(20)   NOT NULL,
    rated_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES User(user_id) ON DELETE CASCADE,
    FOREIGN KEY (movie_id) REFERENCES Movie(movie_id) ON DELETE CASCADE,
    UNIQUE (user_id, movie_id)
);

CREATE TABLE Viewing_History (
    viewing_id        INT           PRIMARY KEY AUTO_INCREMENT,
    user_id           INT           NOT NULL,
    movie_id          INT           NOT NULL,
    watched_date      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completion_status VARCHAR(20)   NOT NULL DEFAULT 'completed',
    watch_count       INT           NOT NULL DEFAULT 1,
    notes             TEXT,
    FOREIGN KEY (user_id) REFERENCES User(user_id) ON DELETE CASCADE,
    FOREIGN KEY (movie_id) REFERENCES Movie(movie_id) ON DELETE CASCADE
);

CREATE TABLE Weekly_Recommendation (
    recommendation_id INT           PRIMARY KEY AUTO_INCREMENT,
    user_id           INT           NOT NULL,
    movie_id          INT           NOT NULL,
    assigned_date     DATE          NOT NULL,
    due_date          DATE          NOT NULL,
    status            VARCHAR(20)   NOT NULL DEFAULT 'pending',
    FOREIGN KEY (user_id) REFERENCES User(user_id) ON DELETE CASCADE,
    FOREIGN KEY (movie_id) REFERENCES Movie(movie_id) ON DELETE CASCADE
);

CREATE TABLE Movie_Contributor (
    movie_id    INT NOT NULL,
    person_id   INT NOT NULL,
    role_id     INT NOT NULL,
    PRIMARY KEY (movie_id, person_id, role_id),
    FOREIGN KEY (movie_id) REFERENCES Movie(movie_id) ON DELETE CASCADE,
    FOREIGN KEY (person_id) REFERENCES Person(person_id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES Role(role_id) ON DELETE CASCADE
);

CREATE TABLE Movie_Genre (
    movie_id  INT NOT NULL,
    genre_id  INT NOT NULL,
    PRIMARY KEY (movie_id, genre_id),
    FOREIGN KEY (movie_id) REFERENCES Movie(movie_id) ON DELETE CASCADE,
    FOREIGN KEY (genre_id) REFERENCES Genre(genre_id) ON DELETE CASCADE
);

CREATE TABLE Movie_Streaming (
    movie_id    INT NOT NULL,
    service_id  INT NOT NULL,
    PRIMARY KEY (movie_id, service_id),
    FOREIGN KEY (movie_id) REFERENCES Movie(movie_id) ON DELETE CASCADE,
    FOREIGN KEY (service_id) REFERENCES Streaming_Service(service_id) ON DELETE CASCADE
);

CREATE TABLE Favorites_Movie (
    favorites_id INT NOT NULL,
    movie_id     INT NOT NULL,
    PRIMARY KEY (favorites_id, movie_id),
    FOREIGN KEY (favorites_id) REFERENCES Favorites_List(favorites_id) ON DELETE CASCADE,
    FOREIGN KEY (movie_id) REFERENCES Movie(movie_id) ON DELETE CASCADE
);

CREATE TABLE Subscription (
    subscription_id    INT           PRIMARY KEY AUTO_INCREMENT,
    user_id            INT           NOT NULL,
    service_id         INT           NOT NULL,
    subscription_status VARCHAR(20)  NOT NULL DEFAULT 'active',
    start_date         DATE          NOT NULL,
    end_date           DATE,
    plan_type          VARCHAR(50),
    auto_renew         BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at         DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES User(user_id) ON DELETE CASCADE,
    FOREIGN KEY (service_id) REFERENCES Streaming_Service(service_id) ON DELETE CASCADE
);

-- ============================================================
-- SAMPLE DATA (20+ tuples in major tables)
-- ============================================================

-- Users (20 rows)
INSERT INTO User (username, email) VALUES
('tom_w',       'tom@example.com'),
('rohan_v',     'rohan@example.com'),
('aanya_a',     'aanya@example.com'),
('ibrahim_s',   'ibrahim@example.com'),
('jane_doe',    'jane@example.com'),
('mike_c',      'mike@example.com'),
('sarah_l',     'sarah@example.com'),
('kevin_p',     'kevin@example.com'),
('lisa_m',      'lisa@example.com'),
('david_r',     'david@example.com'),
('emma_w',      'emma@example.com'),
('james_b',     'james@example.com'),
('olivia_k',    'olivia@example.com'),
('noah_t',      'noah@example.com'),
('sophia_g',    'sophia@example.com'),
('liam_h',      'liam@example.com'),
('ava_j',       'ava@example.com'),
('mason_d',     'mason@example.com'),
('isabella_f',  'isabella@example.com'),
('ethan_z',     'ethan@example.com');

-- Movies (22 rows)
INSERT INTO Movie (title, plot_summary, trailer_url, tmdb_id) VALUES
('The Shawshank Redemption',           'Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.', 'https://youtube.com/watch?v=example1', '278'),
('Inception',                          'A thief who steals corporate secrets through dream-sharing technology is given the task of planting an idea into a CEO''s mind.', 'https://youtube.com/watch?v=example2', '27205'),
('Parasite',                           'Greed and class discrimination threaten the newly formed symbiotic relationship between the wealthy Park family and the destitute Kim clan.', 'https://youtube.com/watch?v=example3', '496243'),
('The Dark Knight',                    'When the menace known as the Joker wreaks havoc on Gotham, Batman must accept one of the greatest tests.', 'https://youtube.com/watch?v=example4', '155'),
('Spirited Away',                      'During her family''s move to the suburbs, a young girl wanders into a world ruled by gods, witches, and spirits.', 'https://youtube.com/watch?v=example5', '129'),
('Whiplash',                           'A promising young drummer enrolls at a cut-throat music conservatory where his dreams of greatness are mentored by an instructor who will stop at nothing.', 'https://youtube.com/watch?v=example6', '244786'),
('Everything Everywhere All at Once',  'An aging Chinese immigrant is swept up in an insane adventure where she alone can save the world by exploring other universes.', 'https://youtube.com/watch?v=example7', '545611'),
('Interstellar',                       'A team of explorers travel through a wormhole in space in an attempt to ensure humanity''s survival.', 'https://youtube.com/watch?v=example8', '157336'),
('The Grand Budapest Hotel',           'A writer encounters the owner of an aging high-class hotel, who tells him of his early years serving as a lobby boy.', 'https://youtube.com/watch?v=example9', '120467'),
('Get Out',                            'A young African-American visits his white girlfriend''s parents for the weekend, where his simmering uneasiness about their reception of him eventually reaches a boiling point.', 'https://youtube.com/watch?v=example10', '419430'),
('La La Land',                         'While navigating their careers in Los Angeles, a pianist and an actress fall in love while attempting to reconcile their aspirations.', 'https://youtube.com/watch?v=example11', '313369'),
('Coco',                               'Aspiring musician Miguel, confronted with his family''s ban on music, enters the Land of the Dead to find his great-great-grandfather.', 'https://youtube.com/watch?v=example12', '354912'),
('Dune',                               'Paul Atreides leads nomadic tribes in a battle to control the desert planet Arrakis.', 'https://youtube.com/watch?v=example13', '438631'),
('The Batman',                         'When a sadistic serial killer begins murdering key political figures in Gotham, Batman is forced to investigate the city''s hidden corruption.', 'https://youtube.com/watch?v=example14', '414906'),
('Oppenheimer',                        'The story of American scientist J. Robert Oppenheimer and his role in the development of the atomic bomb.', 'https://youtube.com/watch?v=example15', '872585'),
('Spider-Man: Across the Spider-Verse','Miles Morales catapults across the Multiverse, where he encounters a team of Spider-People.', 'https://youtube.com/watch?v=example16', '569094'),
('Past Lives',                         'Two childhood friends are separated after one family emigrates from South Korea. Twenty years later they reunite in New York.', 'https://youtube.com/watch?v=example17', '666277'),
('The Menu',                           'A couple travels to a coastal island to eat at an exclusive restaurant where the chef has prepared a lavish menu.', 'https://youtube.com/watch?v=example18', '593643'),
('Aftersun',                           'Sophie reflects on the shared joy and private melancholy of a holiday she took with her father twenty years earlier.', 'https://youtube.com/watch?v=example19', '965150'),
('Knives Out',                         'A detective investigates the death of the patriarch of an eccentric, combative family.', 'https://youtube.com/watch?v=example20', '546554'),
('Moonlight',                          'A young African-American man grapples with his identity and sexuality while experiencing the everyday struggles of childhood, adolescence, and adulthood.', 'https://youtube.com/watch?v=example21', '376867'),
('Drive',                              'A mysterious Hollywood stuntman and mechanic moonlights as a getaway driver and discovers that a contract has been put on him.', 'https://youtube.com/watch?v=example22', '64690');

-- Persons (20 rows)
INSERT INTO Person (full_name, birth_date, biography, imdb_id) VALUES
('Christopher Nolan',   '1970-07-30', 'British-American filmmaker known for complex narratives.', 'nm0634240'),
('Bong Joon-ho',        '1969-09-14', 'South Korean filmmaker and screenwriter.', 'nm0094435'),
('Hayao Miyazaki',      '1941-01-05', 'Japanese animator and filmmaker.', 'nm0594503'),
('Damien Chazelle',     '1985-01-19', 'American director and screenwriter.', 'nm3227090'),
('Leonardo DiCaprio',   '1974-11-11', 'American actor and film producer.', 'nm0000138'),
('Song Kang-ho',        '1967-01-17', 'South Korean actor.', 'nm0814280'),
('Morgan Freeman',      '1937-06-01', 'American actor, director, and narrator.', 'nm0000151'),
('Wes Anderson',        '1969-05-01', 'American filmmaker known for distinctive visual and narrative styles.', 'nm0027572'),
('Jordan Peele',        '1979-02-21', 'American actor, comedian, and filmmaker.', 'nm1443502'),
('Denis Villeneuve',    '1967-10-03', 'Canadian filmmaker known for science fiction epics.', 'nm0898288'),
('Matt Reeves',         '1966-04-27', 'American filmmaker known for franchise blockbusters.', 'nm0716257'),
('Daniel Kwan',         '1988-02-10', 'American filmmaker, one half of the Daniels.', 'nm3718007'),
('Timothee Chalamet',   '1995-12-27', 'American actor known for dramatic roles.', 'nm3154303'),
('Robert Pattinson',    '1986-05-13', 'English actor.', 'nm1500155'),
('Cillian Murphy',      '1976-05-25', 'Irish actor.', 'nm0614165'),
('Michelle Yeoh',       '1962-08-06', 'Malaysian actress known for martial arts and dramatic roles.', 'nm0000706'),
('Greta Lee',           '1983-03-07', 'American actress.', 'nm2901014'),
('Ralph Fiennes',       '1962-12-22', 'English actor.', 'nm0000146'),
('Barry Jenkins',       '1979-11-19', 'American filmmaker.', 'nm1503575'),
('Nicolas Winding Refn','1970-09-29', 'Danish filmmaker.', 'nm0716347');

-- Roles
INSERT INTO Role (role_name, description) VALUES
('Director',        'Primary creative lead of the film'),
('Actor',           'Performs a role in the film'),
('Producer',        'Oversees and finances the film'),
('Screenwriter',    'Writes the screenplay'),
('Cinematographer', 'Manages camera and lighting');

-- Genres (10 rows)
INSERT INTO Genre (genre_name, description) VALUES
('Drama',       'Focused on emotional themes and character development'),
('Sci-Fi',      'Explores futuristic or speculative concepts'),
('Thriller',    'Suspenseful storylines designed to keep viewers on edge'),
('Animation',   'Created using animation techniques'),
('Action',      'Fast-paced sequences and physical feats'),
('Comedy',      'Intended to make the audience laugh'),
('Horror',      'Designed to frighten and unsettle'),
('Romance',     'Centered on romantic relationships'),
('Mystery',     'Revolves around solving a puzzle or crime'),
('Musical',     'Features songs and dance as narrative elements');

-- Streaming Services (8 rows)
INSERT INTO Streaming_Service (service_name) VALUES
('Netflix'),
('Amazon Prime Video'),
('Hulu'),
('HBO Max'),
('Disney+'),
('Apple TV+'),
('Peacock'),
('Paramount+');

-- Notifications (20 rows)
INSERT INTO Notification (user_id, notification_type, message_text, status) VALUES
(1,  'recommendation', 'Your Weekly Watch is ready: Inception!', 'read'),
(2,  'recommendation', 'Your Weekly Watch is ready: Parasite!', 'unread'),
(3,  'recommendation', 'Your Weekly Watch is ready: Spirited Away!', 'unread'),
(4,  'reminder',       'Don''t forget to watch your Weekly Watch this week!', 'unread'),
(5,  'recommendation', 'Your Weekly Watch is ready: Whiplash!', 'read'),
(6,  'recommendation', 'Your Weekly Watch is ready: Interstellar!', 'unread'),
(7,  'recommendation', 'Your Weekly Watch is ready: Get Out!', 'read'),
(8,  'recommendation', 'Your Weekly Watch is ready: Dune!', 'unread'),
(9,  'reminder',       'You have 2 days left for your Weekly Watch!', 'unread'),
(10, 'recommendation', 'Your Weekly Watch is ready: La La Land!', 'read'),
(11, 'recommendation', 'Your Weekly Watch is ready: Coco!', 'unread'),
(12, 'recommendation', 'Your Weekly Watch is ready: Oppenheimer!', 'read'),
(13, 'reminder',       'Rate your last Weekly Watch!', 'unread'),
(14, 'recommendation', 'Your Weekly Watch is ready: Knives Out!', 'unread'),
(15, 'recommendation', 'Your Weekly Watch is ready: Moonlight!', 'read'),
(16, 'recommendation', 'Your Weekly Watch is ready: Drive!', 'unread'),
(17, 'recommendation', 'Your Weekly Watch is ready: Past Lives!', 'read'),
(18, 'reminder',       'Your current recommendation expires tomorrow!', 'unread'),
(19, 'recommendation', 'Your Weekly Watch is ready: The Menu!', 'unread'),
(20, 'recommendation', 'Your Weekly Watch is ready: Aftersun!', 'read');

-- User Preferences (20 rows)
INSERT INTO User_Preference (user_id, preference_type, preference_value, strength) VALUES
(1,  'genre',    'Drama',              'strong'),
(1,  'genre',    'Sci-Fi',             'medium'),
(2,  'genre',    'Thriller',           'strong'),
(3,  'genre',    'Animation',          'strong'),
(4,  'director', 'Christopher Nolan',  'strong'),
(5,  'genre',    'Drama',              'medium'),
(6,  'genre',    'Sci-Fi',             'strong'),
(7,  'genre',    'Horror',             'strong'),
(8,  'genre',    'Action',             'medium'),
(9,  'director', 'Denis Villeneuve',   'strong'),
(10, 'genre',    'Romance',            'strong'),
(11, 'genre',    'Animation',          'medium'),
(12, 'genre',    'Drama',              'strong'),
(13, 'genre',    'Mystery',            'strong'),
(14, 'director', 'Wes Anderson',       'medium'),
(15, 'genre',    'Thriller',           'strong'),
(16, 'genre',    'Action',             'strong'),
(17, 'genre',    'Drama',              'medium'),
(18, 'genre',    'Comedy',             'strong'),
(19, 'director', 'Jordan Peele',       'medium');

-- Favorites Lists (20 rows)
INSERT INTO Favorites_List (user_id, list_name) VALUES
(1,  'All-Time Favorites'),
(2,  'Must Rewatch'),
(3,  'Studio Ghibli Collection'),
(4,  'Nolan Films'),
(5,  'Best of 2020s'),
(6,  'Sci-Fi Essentials'),
(7,  'Horror Nights'),
(8,  'Action Packed'),
(9,  'Villeneuve Collection'),
(10, 'Date Night Picks'),
(11, 'Family Favorites'),
(12, 'Award Winners'),
(13, 'Mystery Marathon'),
(14, 'Quirky Picks'),
(15, 'Edge of My Seat'),
(16, 'Superhero Films'),
(17, 'Indie Gems'),
(18, 'Feel Good Movies'),
(19, 'Mind Benders'),
(20, 'Modern Classics');

-- Reviews (20 rows)
INSERT INTO Review (user_id, movie_id, review_text, is_spoiler) VALUES
(1,  1,  'An absolute masterpiece of storytelling. The pacing is perfect.', FALSE),
(2,  3,  'Brilliantly crafted social commentary. Every scene has purpose.', FALSE),
(3,  5,  'A magical journey that appeals to all ages. Miyazaki at his best.', FALSE),
(1,  2,  'Mind-bending plot that rewards multiple viewings.', TRUE),
(4,  4,  'Heath Ledger''s performance is legendary. Best superhero film ever made.', FALSE),
(5,  6,  'Intense and gripping. J.K. Simmons is terrifying and brilliant.', FALSE),
(6,  8,  'Nolan outdoes himself. The docking scene is breathtaking.', FALSE),
(7,  10, 'A masterclass in building tension through social commentary.', FALSE),
(8,  13, 'Epic world-building. Villeneuve captures the scale perfectly.', FALSE),
(9,  15, 'A haunting portrait of ambition and consequence.', FALSE),
(10, 11, 'Beautiful cinematography and music. Bittersweet and romantic.', FALSE),
(11, 12, 'Made me cry. The family themes are so heartfelt.', FALSE),
(12, 7,  'Unlike anything I have ever seen. Absolute creative chaos.', FALSE),
(13, 20, 'Clever mystery with razor-sharp dialogue.', FALSE),
(14, 9,  'Wes Anderson at his most whimsical and charming.', FALSE),
(15, 14, 'Dark and moody. A fresh take on the character.', FALSE),
(16, 16, 'Visually stunning animation. Better than the first.', FALSE),
(17, 17, 'Quiet and devastating. Lingers in your mind for days.', FALSE),
(18, 18, 'Darkly funny with an unforgettable final act.', FALSE),
(19, 21, 'Profound and poetic filmmaking. Every frame is beautiful.', FALSE);

-- Ratings (24 rows)
INSERT INTO Rating (user_id, movie_id, rating_value) VALUES
(1,  1,  'loved'),
(1,  2,  'loved'),
(2,  3,  'loved'),
(3,  5,  'loved'),
(4,  4,  'loved'),
(5,  6,  'liked'),
(2,  1,  'liked'),
(1,  4,  'liked'),
(6,  8,  'loved'),
(7,  10, 'loved'),
(8,  13, 'loved'),
(9,  15, 'loved'),
(10, 11, 'loved'),
(11, 12, 'loved'),
(12, 7,  'loved'),
(13, 20, 'liked'),
(14, 9,  'loved'),
(15, 14, 'liked'),
(16, 16, 'loved'),
(17, 17, 'loved'),
(18, 18, 'liked'),
(19, 21, 'loved'),
(20, 22, 'liked'),
(1,  7,  'loved');

-- Viewing History (22 rows)
INSERT INTO Viewing_History (user_id, movie_id, watched_date, completion_status, watch_count, notes) VALUES
(1,  1,  '2026-01-15 20:00:00', 'completed', 3, 'Third rewatch, still amazing'),
(1,  2,  '2026-01-22 19:30:00', 'completed', 2, 'Caught new details second time'),
(2,  3,  '2026-02-01 21:00:00', 'completed', 1, 'Watched for first time, blown away'),
(3,  5,  '2026-02-08 18:00:00', 'completed', 5, 'My comfort movie'),
(4,  4,  '2026-02-10 20:30:00', 'completed', 2, NULL),
(5,  6,  '2026-02-15 19:00:00', 'completed', 1, 'Recommended by a friend'),
(1,  7,  '2026-03-01 20:00:00', 'completed', 1, 'What a ride'),
(6,  8,  '2026-01-10 21:00:00', 'completed', 2, 'The docking scene is unreal'),
(7,  10, '2026-01-18 20:00:00', 'completed', 1, 'So unsettling and smart'),
(8,  13, '2026-02-05 19:30:00', 'completed', 1, 'Incredible world building'),
(9,  15, '2026-02-20 20:00:00', 'completed', 1, 'Heavy but brilliant'),
(10, 11, '2026-01-25 19:00:00', 'completed', 3, 'My favorite romance film'),
(11, 12, '2026-02-12 18:00:00', 'completed', 2, 'Kids loved it too'),
(12, 7,  '2026-03-05 21:00:00', 'completed', 1, 'Absolutely wild'),
(13, 20, '2026-01-30 20:30:00', 'completed', 1, 'Great twist ending'),
(14, 9,  '2026-02-18 19:00:00', 'completed', 1, 'Such a charming film'),
(15, 14, '2026-03-02 21:00:00', 'completed', 1, 'Loved the noir atmosphere'),
(16, 16, '2026-02-25 18:30:00', 'completed', 2, 'Animation is next level'),
(17, 17, '2026-03-08 19:00:00', 'completed', 1, 'Beautifully understated'),
(18, 18, '2026-03-10 20:00:00', 'completed', 1, 'Unexpected and thrilling'),
(19, 21, '2026-03-12 19:30:00', 'completed', 1, 'Emotionally devastating'),
(20, 22, '2026-03-15 21:00:00', 'completed', 1, 'Cool and stylish');

-- Weekly Recommendations (20 rows)
INSERT INTO Weekly_Recommendation (user_id, movie_id, assigned_date, due_date, status) VALUES
(1,  2,  '2026-01-20', '2026-01-27', 'completed'),
(2,  3,  '2026-02-01', '2026-02-08', 'completed'),
(3,  5,  '2026-02-08', '2026-02-15', 'completed'),
(4,  6,  '2026-02-15', '2026-02-22', 'pending'),
(5,  7,  '2026-03-01', '2026-03-08', 'pending'),
(1,  3,  '2026-03-10', '2026-03-17', 'pending'),
(6,  8,  '2026-01-06', '2026-01-13', 'completed'),
(7,  10, '2026-01-13', '2026-01-20', 'completed'),
(8,  13, '2026-02-01', '2026-02-08', 'completed'),
(9,  15, '2026-02-15', '2026-02-22', 'completed'),
(10, 11, '2026-01-20', '2026-01-27', 'completed'),
(11, 12, '2026-02-08', '2026-02-15', 'completed'),
(12, 7,  '2026-03-01', '2026-03-08', 'completed'),
(13, 20, '2026-01-27', '2026-02-03', 'completed'),
(14, 9,  '2026-02-15', '2026-02-22', 'completed'),
(15, 14, '2026-02-27', '2026-03-06', 'completed'),
(16, 16, '2026-02-22', '2026-03-01', 'completed'),
(17, 17, '2026-03-05', '2026-03-12', 'pending'),
(18, 18, '2026-03-08', '2026-03-15', 'pending'),
(19, 21, '2026-03-10', '2026-03-17', 'pending');

-- Movie_Contributor (bridge table)
INSERT INTO Movie_Contributor (movie_id, person_id, role_id) VALUES
(2,  1,  1),
(3,  2,  1),
(5,  3,  1),
(6,  4,  1),
(2,  5,  2),
(3,  6,  2),
(1,  7,  2),
(9,  8,  1),
(10, 9,  1),
(13, 10, 1),
(14, 11, 1),
(7,  12, 1),
(13, 13, 2),
(14, 14, 2),
(15, 1,  1),
(15, 15, 2),
(7,  16, 2),
(17, 17, 2),
(9,  18, 2),
(21, 19, 1),
(22, 20, 1),
(8,  1,  1),
(11, 4,  1);

-- Movie_Genre (bridge table)
INSERT INTO Movie_Genre (movie_id, genre_id) VALUES
(1,  1),
(2,  2),
(2,  5),
(3,  1),
(3,  3),
(4,  5),
(4,  3),
(5,  4),
(6,  1),
(7,  2),
(7,  5),
(7,  6),
(8,  2),
(8,  1),
(9,  6),
(10, 7),
(10, 3),
(11, 8),
(11, 10),
(12, 4),
(13, 2),
(13, 5),
(14, 5),
(14, 3),
(15, 1),
(16, 4),
(16, 5),
(17, 1),
(17, 8),
(18, 3),
(18, 6),
(19, 1),
(20, 9),
(20, 3),
(21, 1),
(22, 5),
(22, 3);

-- Movie_Streaming (bridge table)
INSERT INTO Movie_Streaming (movie_id, service_id) VALUES
(1,  1),
(2,  1),
(2,  4),
(3,  3),
(4,  4),
(5,  4),
(6,  2),
(7,  6),
(8,  8),
(9,  5),
(10, 7),
(11, 1),
(12, 5),
(13, 4),
(14, 4),
(15, 7),
(16, 1),
(17, 6),
(18, 4),
(20, 2),
(21, 1),
(22, 3);

-- Favorites_Movie (bridge table)
INSERT INTO Favorites_Movie (favorites_id, movie_id) VALUES
(1,  1),
(1,  2),
(2,  3),
(3,  5),
(4,  4),
(4,  2),
(5,  7),
(6,  8),
(6,  2),
(7,  10),
(8,  13),
(8,  4),
(9,  13),
(10, 11),
(11, 12),
(11, 5),
(12, 7),
(12, 15),
(13, 20),
(14, 9);

-- Subscription (20 rows)
INSERT INTO Subscription (user_id, service_id, subscription_status, start_date, end_date, plan_type, auto_renew) VALUES
(1,  1, 'active',    '2025-06-01', NULL,         'Premium',  TRUE),
(1,  4, 'active',    '2025-09-01', NULL,         'Standard', TRUE),
(2,  3, 'active',    '2025-07-15', NULL,         'Basic',    TRUE),
(3,  4, 'active',    '2025-08-01', NULL,         'Premium',  TRUE),
(4,  1, 'cancelled', '2025-01-01', '2025-12-31', 'Standard', FALSE),
(5,  2, 'active',    '2026-01-01', NULL,         'Premium',  TRUE),
(5,  6, 'active',    '2026-02-01', NULL,         'Standard', TRUE),
(6,  1, 'active',    '2025-10-01', NULL,         'Premium',  TRUE),
(7,  7, 'active',    '2025-11-01', NULL,         'Basic',    TRUE),
(8,  4, 'active',    '2025-12-01', NULL,         'Standard', TRUE),
(9,  8, 'active',    '2026-01-15', NULL,         'Premium',  TRUE),
(10, 1, 'active',    '2025-09-15', NULL,         'Standard', TRUE),
(11, 5, 'active',    '2025-08-15', NULL,         'Premium',  TRUE),
(12, 6, 'active',    '2026-01-01', NULL,         'Standard', TRUE),
(13, 2, 'active',    '2025-07-01', NULL,         'Basic',    TRUE),
(14, 5, 'active',    '2025-11-15', NULL,         'Premium',  TRUE),
(15, 4, 'active',    '2026-02-01', NULL,         'Standard', TRUE),
(16, 1, 'active',    '2026-01-10', NULL,         'Premium',  TRUE),
(17, 6, 'active',    '2026-02-15', NULL,         'Standard', TRUE),
(18, 4, 'cancelled', '2025-06-01', '2026-01-01', 'Basic',    FALSE);