-- ============================================================
-- The Weekly Watch: A Movie Recommendation System
-- Phase 4: Normalized Relational Schema and Database Initialization
-- CS 4604 - Spring 2026
-- DBMS: MySQL 8.x
-- ============================================================

DROP DATABASE IF EXISTS weekly_watch;
CREATE DATABASE weekly_watch;
USE weekly_watch;

-- ============================================================
-- TABLE DEFINITIONS (ordered by dependency)
-- ============================================================

-- Independent entities first (no foreign keys)

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

-- Dependent entities (have foreign keys to above tables)

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

-- Bridge / associative tables

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
-- SAMPLE DATA (minimum 5 tuples per major table)
-- ============================================================

-- Users
INSERT INTO User (username, email) VALUES
('tom_w',      'tom@example.com'),
('rohan_v',    'rohan@example.com'),
('aanya_a',    'aanya@example.com'),
('ibrahim_s',  'ibrahim@example.com'),
('jane_doe',   'jane@example.com'),
('mike_c',     'mike@example.com');

-- Movies
INSERT INTO Movie (title, plot_summary, trailer_url, tmdb_id) VALUES
('The Shawshank Redemption', 'Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.', 'https://youtube.com/watch?v=example1', '278'),
('Inception',                'A thief who steals corporate secrets through dream-sharing technology is given the task of planting an idea into a CEO''s mind.', 'https://youtube.com/watch?v=example2', '27205'),
('Parasite',                 'Greed and class discrimination threaten the newly formed symbiotic relationship between the wealthy Park family and the destitute Kim clan.', 'https://youtube.com/watch?v=example3', '496243'),
('The Dark Knight',          'When the menace known as the Joker wreaks havoc on Gotham, Batman must accept one of the greatest tests.', 'https://youtube.com/watch?v=example4', '155'),
('Spirited Away',            'During her family''s move to the suburbs, a young girl wanders into a world ruled by gods, witches, and spirits.', 'https://youtube.com/watch?v=example5', '129'),
('Whiplash',                 'A promising young drummer enrolls at a cut-throat music conservatory where his dreams of greatness are mentored by an instructor who will stop at nothing.', 'https://youtube.com/watch?v=example6', '244786'),
('Everything Everywhere All at Once', 'An aging Chinese immigrant is swept up in an insane adventure where she alone can save the world by exploring other universes.', 'https://youtube.com/watch?v=example7', '545611');

-- Persons (directors and actors)
INSERT INTO Person (full_name, birth_date, biography, imdb_id) VALUES
('Christopher Nolan',  '1970-07-30', 'British-American filmmaker known for complex narratives.', 'nm0634240'),
('Bong Joon-ho',       '1969-09-14', 'South Korean filmmaker and screenwriter.', 'nm0094435'),
('Hayao Miyazaki',     '1941-01-05', 'Japanese animator and filmmaker.', 'nm0594503'),
('Damien Chazelle',    '1985-01-19', 'American director and screenwriter.', 'nm3227090'),
('Leonardo DiCaprio',  '1974-11-11', 'American actor and film producer.', 'nm0000138'),
('Song Kang-ho',       '1967-01-17', 'South Korean actor.', 'nm0814280'),
('Morgan Freeman',     '1937-06-01', 'American actor, director, and narrator.', 'nm0000151');

-- Roles
INSERT INTO Role (role_name, description) VALUES
('Director',       'Primary creative lead of the film'),
('Actor',          'Performs a role in the film'),
('Producer',       'Oversees and finances the film'),
('Screenwriter',   'Writes the screenplay'),
('Cinematographer', 'Manages camera and lighting');

-- Genres
INSERT INTO Genre (genre_name, description) VALUES
('Drama',       'Focused on emotional themes and character development'),
('Sci-Fi',      'Explores futuristic or speculative concepts'),
('Thriller',    'Suspenseful storylines designed to keep viewers on edge'),
('Animation',   'Created using animation techniques'),
('Action',      'Fast-paced sequences and physical feats'),
('Comedy',      'Intended to make the audience laugh'),
('Horror',      'Designed to frighten and unsettle');

-- Streaming Services
INSERT INTO Streaming_Service (service_name) VALUES
('Netflix'),
('Amazon Prime Video'),
('Hulu'),
('HBO Max'),
('Disney+'),
('Apple TV+');

-- Notifications
INSERT INTO Notification (user_id, notification_type, message_text, status) VALUES
(1, 'recommendation', 'Your Weekly Watch is ready: Inception!', 'read'),
(2, 'recommendation', 'Your Weekly Watch is ready: Parasite!', 'unread'),
(3, 'recommendation', 'Your Weekly Watch is ready: Spirited Away!', 'unread'),
(4, 'reminder',       'Don''t forget to watch your Weekly Watch this week!', 'unread'),
(5, 'recommendation', 'Your Weekly Watch is ready: Whiplash!', 'read');

-- User Preferences
INSERT INTO User_Preference (user_id, preference_type, preference_value, strength) VALUES
(1, 'genre',    'Drama',     'strong'),
(1, 'genre',    'Sci-Fi',    'medium'),
(2, 'genre',    'Thriller',  'strong'),
(3, 'genre',    'Animation', 'strong'),
(4, 'director', 'Christopher Nolan', 'strong'),
(5, 'genre',    'Drama',     'medium');

-- Favorites Lists
INSERT INTO Favorites_List (user_id, list_name) VALUES
(1, 'All-Time Favorites'),
(2, 'Must Rewatch'),
(3, 'Studio Ghibli Collection'),
(4, 'Nolan Films'),
(5, 'Best of 2020s');

-- Reviews
INSERT INTO Review (user_id, movie_id, review_text, is_spoiler) VALUES
(1, 1, 'An absolute masterpiece of storytelling. The pacing is perfect.', FALSE),
(2, 3, 'Brilliantly crafted social commentary. Every scene has purpose.', FALSE),
(3, 5, 'A magical journey that appeals to all ages. Miyazaki at his best.', FALSE),
(1, 2, 'Mind-bending plot that rewards multiple viewings. The ending still gets me.', TRUE),
(4, 4, 'Heath Ledger''s performance is legendary. Best superhero film ever made.', FALSE),
(5, 6, 'Intense and gripping. J.K. Simmons is terrifying and brilliant.', FALSE);

-- Ratings
INSERT INTO Rating (user_id, movie_id, rating_value) VALUES
(1, 1, 'loved'),
(1, 2, 'loved'),
(2, 3, 'loved'),
(3, 5, 'loved'),
(4, 4, 'loved'),
(5, 6, 'liked'),
(2, 1, 'liked'),
(1, 4, 'liked');

-- Viewing History
INSERT INTO Viewing_History (user_id, movie_id, watched_date, completion_status, watch_count, notes) VALUES
(1, 1, '2026-01-15 20:00:00', 'completed', 3, 'Third rewatch, still amazing'),
(1, 2, '2026-01-22 19:30:00', 'completed', 2, 'Caught new details second time'),
(2, 3, '2026-02-01 21:00:00', 'completed', 1, 'Watched for first time, blown away'),
(3, 5, '2026-02-08 18:00:00', 'completed', 5, 'My comfort movie'),
(4, 4, '2026-02-10 20:30:00', 'completed', 2, NULL),
(5, 6, '2026-02-15 19:00:00', 'completed', 1, 'Recommended by a friend'),
(1, 7, '2026-03-01 20:00:00', 'completed', 1, 'What a ride');

-- Weekly Recommendations
INSERT INTO Weekly_Recommendation (user_id, movie_id, assigned_date, due_date, status) VALUES
(1, 2, '2026-01-20', '2026-01-27', 'completed'),
(2, 3, '2026-02-01', '2026-02-08', 'completed'),
(3, 5, '2026-02-08', '2026-02-15', 'completed'),
(4, 6, '2026-02-15', '2026-02-22', 'pending'),
(5, 7, '2026-03-01', '2026-03-08', 'pending'),
(1, 3, '2026-03-10', '2026-03-17', 'pending');

-- Movie_Contributor (bridge table)
INSERT INTO Movie_Contributor (movie_id, person_id, role_id) VALUES
(2, 1, 1),  -- Nolan directed Inception
(3, 2, 1),  -- Bong Joon-ho directed Parasite
(5, 3, 1),  -- Miyazaki directed Spirited Away
(6, 4, 1),  -- Chazelle directed Whiplash
(2, 5, 2),  -- DiCaprio acted in Inception
(3, 6, 2),  -- Song Kang-ho acted in Parasite
(1, 7, 2);  -- Morgan Freeman acted in Shawshank

-- Movie_Genre (bridge table)
INSERT INTO Movie_Genre (movie_id, genre_id) VALUES
(1, 1),  -- Shawshank - Drama
(2, 2),  -- Inception - Sci-Fi
(2, 5),  -- Inception - Action
(3, 1),  -- Parasite - Drama
(3, 3),  -- Parasite - Thriller
(4, 5),  -- Dark Knight - Action
(4, 3),  -- Dark Knight - Thriller
(5, 4),  -- Spirited Away - Animation
(6, 1),  -- Whiplash - Drama
(7, 2),  -- EEAAO - Sci-Fi
(7, 5),  -- EEAAO - Action
(7, 6);  -- EEAAO - Comedy

-- Movie_Streaming (bridge table)
INSERT INTO Movie_Streaming (movie_id, service_id) VALUES
(1, 1),  -- Shawshank on Netflix
(2, 1),  -- Inception on Netflix
(2, 4),  -- Inception on HBO Max
(3, 3),  -- Parasite on Hulu
(4, 4),  -- Dark Knight on HBO Max
(5, 4),  -- Spirited Away on HBO Max
(6, 2),  -- Whiplash on Amazon Prime
(7, 6);  -- EEAAO on Apple TV+

-- Favorites_Movie (bridge table)
INSERT INTO Favorites_Movie (favorites_id, movie_id) VALUES
(1, 1),  -- Tom's favorites: Shawshank
(1, 2),  -- Tom's favorites: Inception
(2, 3),  -- Rohan's rewatch: Parasite
(3, 5),  -- Aanya's Ghibli: Spirited Away
(4, 4),  -- Ibrahim's Nolan: Dark Knight
(4, 2),  -- Ibrahim's Nolan: Inception
(5, 7);  -- Jane's 2020s: EEAAO

-- Subscription (bridge table)
INSERT INTO Subscription (user_id, service_id, subscription_status, start_date, end_date, plan_type, auto_renew) VALUES
(1, 1, 'active',   '2025-06-01', NULL,          'Premium',  TRUE),
(1, 4, 'active',   '2025-09-01', NULL,          'Standard', TRUE),
(2, 3, 'active',   '2025-07-15', NULL,          'Basic',    TRUE),
(3, 4, 'active',   '2025-08-01', NULL,          'Premium',  TRUE),
(4, 1, 'cancelled','2025-01-01', '2025-12-31',  'Standard', FALSE),
(5, 2, 'active',   '2026-01-01', NULL,          'Premium',  TRUE),
(5, 6, 'active',   '2026-02-01', NULL,          'Standard', TRUE);
