-- CREATE DATABASE test OWNER postgres ENCODING 'UTF8' LC_COLLATE 'English_United States.1252' LC_CTYPE 'English_United States.1252';
-- CREATE DATABASE test WITH OWNER "postgres" ENCODING 'UTF8' LC_COLLATE='en_US.UTF-8' LC_CTYPE='en_US.UTF-8' TEMPLATE template0;
-- CREATE TABLE Users (
--     about TEXT NOT NULL DEFAULT '',
--     email VARCHAR(255) NOT NULL PRIMARY KEY CONSTRAINT email_right CHECK('^.*@[A-Za-z0-9\-_\.]*$'),
--     fullname TEXT NOT NULL DEFAULT '',
--     nickname CITEXT COLLATE "English_United States.1252" CONSTRAINT nick_right CHECK(nickname'^[A-Za-z0-9]*$') UNIQUE
-- );
DROP TABLE IF EXISTS Vote;
DROP TABLE IF EXISTS Post;
DROP TABLE IF EXISTS Thread;
DROP TABLE IF EXISTS Forum;
DROP TABLE IF EXISTS Users;
CREATE TABLE Users (
    about TEXT NOT NULL DEFAULT '',
    email VARCHAR(255) NOT NULL PRIMARY KEY CONSTRAINT email_right CHECK(email ~ '^.*@[A-Za-z0-9\-_\.]*$'),
    fullname TEXT NOT NULL DEFAULT '',
    nickname CITEXT COLLATE pg_catalog."en-US-x-icu" CONSTRAINT nick_right CHECK(nickname ~ '^[A-Za-z0-9]*$') UNIQUE
);

-- DROP TABLE IF EXISTS Forum;
CREATE TABLE Forum (
    posts BIGINT CONSTRAINT non_negative_posts_count CHECK (posts>=0) NOT NULL DEFAULT 0,  --autoincrement
    slug TEXT PRIMARY KEY UNIQUE CONSTRAINT slug_correct CHECK(slug ~ '^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$'),
    threads INTEGER CONSTRAINT non_negative_threads_count CHECK (threads>=0) DEFAULT 0,
    title TEXT NOT NULL DEFAULT '',
    userNick CITEXT REFERENCES Users (nickname) ON DELETE RESTRICT ON UPDATE RESTRICT NOT NULL
);


CREATE OR REPLACE FUNCTION slug_thread() RETURNS TEXT LANGUAGE SQL AS
$$ SELECT array_to_string(array(select substr('ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghigklmnopqrstuvwxyz0123456789-_',((random()*(62-1)+1)::integer),1) from generate_series(1,10)),'') $$;

-- DROP TABLE IF EXISTS Thread;
CREATE TABLE Thread (
    author CITEXT REFERENCES Users (nickname) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    forum TEXT REFERENCES Forum (slug) ON DELETE CASCADE ON UPDATE RESTRICT,
    id BIGSERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    slug TEXT UNIQUE CONSTRAINT slug_correct CHECK(slug ~ '^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$') DEFAULT slug_thread(),
    title TEXT NOT NULL,
    votes INTEGER NOT NULL DEFAULT 0
);


-- DROP TABLE IF EXISTS Post;
CREATE TABLE Post (
    author CITEXT REFERENCES Users (nickname) ON DELETE RESTRICT ON UPDATE RESTRICT NOT NULL,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    forum TEXT REFERENCES Forum (slug) ON DELETE CASCADE ON UPDATE RESTRICT,
    id BIGSERIAL PRIMARY KEY,
    isEdited BOOLEAN NOT NULL DEFAULT false,
    message TEXT NOT NULL DEFAULT '',
    parent BIGINT NOT NULL DEFAULT 0,
    thread INTEGER REFERENCES Thread (id) ON DELETE CASCADE ON UPDATE RESTRICT
);


CREATE TABLE vote (
    thread BIGINT REFERENCES Thread (id) ON DELETE CASCADE ON UPDATE CASCADE,
    author CITEXT REFERENCES Users (nickname) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL,
    UNIQUE (thread, author)
);

------------Триггеры-------------------------

-- Триггер на User. Отвечает за read-only поля.
CREATE OR REPLACE FUNCTION user_readonly() RETURNS trigger AS $user_readonly$
    BEGIN
        IF NEW.nickname!=OLD.nickname THEN
            RAISE EXCEPTION 'read-only (.nickname)';
        END IF;
        RETURN NEW;
    END
$user_readonly$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS user_readonly ON Users;
CREATE TRIGGER user_readonly BEFORE UPDATE ON Users
    FOR EACH ROW EXECUTE PROCEDURE user_readonly();


-- Триггер на Post-ы. Отвечает за счетчики в Forum и за read-only данные Post
CREATE OR REPLACE FUNCTION update_forum_posts() RETURNS trigger AS $update_forum_posts$
    BEGIN
        IF TG_OP='INSERT' THEN
            UPDATE Forum SET posts=posts+1 WHERE slug=NEW.forum;
            RETURN NEW;
        ELSIF TG_OP='DELETE' OR TG_OP='TRUNCATE' THEN
            UPDATE Forum SET posts=posts-1 WHERE slug=OLD.forum;
            RETURN OLD;
        ELSE
            IF NEW.created!=OLD.created THEN
                RAISE EXCEPTION 'const .created';
            END IF;
            IF NEW.forum!=OLD.forum THEN
                RAISE EXCEPTION 'const .forum';
            END IF;
            IF NEW.id!=OLD.id THEN
                RAISE EXCEPTION 'const .id';
            END IF;
            IF NEW.isEdited!=OLD.isEdited THEN
                RAISE EXCEPTION 'const .isEdited';
            END IF;
            IF NEW.thread!=OLD.thread THEN
                RAISE EXCEPTION 'const .thread';
            END IF;
            New.isEdited='true';
        END IF;
    END
$update_forum_posts$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS update_forum_posts ON Post;
CREATE TRIGGER update_forum_posts AFTER UPDATE OR INSERT OR DELETE ON Post
    FOR EACH ROW EXECUTE PROCEDURE update_forum_posts();


-- Триггер на Thread-ы. Отвечает за счетчик в Forums и const-поля
CREATE OR REPLACE FUNCTION update_forum_threads() RETURNS trigger AS $update_forum_threads$
    BEGIN
        RAISE NOTICE 'forum.threads to update';
        IF TG_OP='INSERT' THEN
            UPDATE Forum SET threads=threads+1 WHERE slug=NEW.forum;
            RETURN NEW;
        ELSIF TG_OP='DELETE' OR TG_OP='TRUNCATE' THEN
            UPDATE Forum SET threads=threads-1 WHERE slug=OLD.forum;
            RETURN OLD;
        ELSE
            IF NEW.forum!=OLD.forum THEN
                RAISE EXCEPTION 'const .forum';
            END IF;
            IF NEW.id!=OLD.id THEN
                RAISE EXCEPTION 'const .id';
            END IF;
            IF NEW.slug!=OLD.slug THEN
                RAISE EXCEPTION 'const .slug';
            END IF;
            RETURN NEW;
        END IF;
    END
$update_forum_threads$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS update_forum_threads ON Thread;
CREATE TRIGGER update_forum_threads AFTER UPDATE OR INSERT OR DELETE ON Thread
    FOR EACH ROW EXECUTE PROCEDURE update_forum_threads();

-- Триггер на Vote-ы. Отвечает за счетчики votes у Thread-ов.
CREATE OR REPLACE FUNCTION update_thread_vote_counter() RETURNS trigger AS $update_thread_vote_counter$
    BEGIN
        IF TG_OP='INSERT' THEN
            UPDATE Thread SET votes=votes+1 WHERE id=NEW.thread;
            RETURN NEW;
        ELSE -- DELETE
            UPDATE Thread SET votes=votes-1 WHERE id=OLD.thread;
            RETURN OLD;
        END IF;
    END
$update_thread_vote_counter$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS update_thread_vote ON Vote;
CREATE TRIGGER update_thread_vote AFTER INSERT OR DELETE ON Vote
    FOR EACH ROW EXECUTE PROCEDURE update_thread_vote_counter();

INSERT INTO Users (email, fullname, nickname, about) VALUES ('ivanov.vanya@mail.ry', 'Ian', 'tamerlanchik', 'About me');
INSERT INTO Forum (slug, title, userNick) VALUES ('test_forum', 'Hello, world', 'tamerlanchik');
INSERT INTO Thread (author, forum, message, slug, title) VALUES ('tamerlanchik', 'test_forum', 'Hello fucking world!', 'test_thread', 'Fucking world!'), ('tamerlanchik', 'test_forum', 'Die', 'next_thread', 'DieDie');
INSERT INTO Post (author, forum, message, thread) VALUES ('tamerlanchik', 'test_forum', 'Hello', 1);
