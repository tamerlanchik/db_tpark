-- CREATE DATABASE test OWNER postgres ENCODING 'UTF8' LC_COLLATE 'English_United States.1252' LC_CTYPE 'English_United States.1252';
-- CREATE DATABASE test WITH OWNER "postgres" ENCODING 'UTF8' LC_COLLATE='en_US.UTF-8' LC_CTYPE='en_US.UTF-8' TEMPLATE template0;
-- CREATE TABLE Users (
--     about TEXT NOT NULL DEFAULT '',
--     email VARCHAR(255) NOT NULL PRIMARY KEY CONSTRAINT email_right CHECK('^.*@[A-Za-z0-9\-_\.]*$'),
--     fullname TEXT NOT NULL DEFAULT '',
--     nickname CITEXT COLLATE "English_United States.1252" CONSTRAINT nick_right CHECK(nickname'^[A-Za-z0-9]*$') UNIQUE
-- );
CREATE EXTENSION citext;
DROP TABLE IF EXISTS Vote;
DROP TABLE IF EXISTS Post;
DROP TABLE IF EXISTS Thread;
DROP TABLE IF EXISTS Forum;
DROP TABLE IF EXISTS Users;


-- _______Users________
CREATE TABLE Users (
    about TEXT NOT NULL DEFAULT '',
    email CITEXT NOT NULL UNIQUE CONSTRAINT email_right CHECK(email ~ '^.*@[A-Za-z0-9\-_\.]*$'),
    fullname TEXT NOT NULL DEFAULT '',
--     nickname CITEXT COLLATE pg_catalog."en-US-x-icu" PRIMARY KEY CONSTRAINT nick_right CHECK(nickname ~ '^[A-Za-z0-9_\.]*$')
    nickname CITEXT COLLATE "POSIX" PRIMARY KEY CONSTRAINT nick_right CHECK(nickname ~ '^[A-Za-z0-9_\.]*$')
);
CREATE UNIQUE INDEX users_nickname_index on Users (LOWER(nickname));

-- _________Forum____________
CREATE TABLE Forum (
    posts BIGINT CONSTRAINT non_negative_posts_count CHECK (posts>=0) NOT NULL DEFAULT 0,  --autoincrement
    slug CITEXT PRIMARY KEY UNIQUE CONSTRAINT slug_correct CHECK(slug ~ '^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$'),
    threads INTEGER CONSTRAINT non_negative_threads_count CHECK (threads>=0) DEFAULT 0,
    title TEXT NOT NULL DEFAULT '',
    userNick CITEXT REFERENCES Users (nickname) ON DELETE RESTRICT ON UPDATE RESTRICT NOT NULL
);
CREATE UNIQUE INDEX forum_slug_index on Forum (LOWER(slug));

-- _______Thread__________

-- CREATE OR REPLACE FUNCTION slug_thread() RETURNS TEXT LANGUAGE SQL AS
-- $$ SELECT array_to_string(array(select substr('ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghigklmnopqrstuvwxyz0123456789-_',((random()*(62-1)+1)::integer),1) from generate_series(1,10)),'') $$;

CREATE TABLE Thread (
    author CITEXT REFERENCES Users (nickname) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    forum CITEXT REFERENCES Forum (slug) ON DELETE CASCADE ON UPDATE RESTRICT,
    id BIGSERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    slug CITEXT UNIQUE CONSTRAINT slug_correct CHECK(slug ~ '^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$'),
    title TEXT NOT NULL,
    votes INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX thread_slug_index on Thread (LOWER(slug));

-- _______Post___________
CREATE OR REPLACE FUNCTION get_thread_by_post(post_ BIGINT) RETURNS INTEGER AS $get_post_thread$
    BEGIN
        RETURN (SELECT thread FROM Post WHERE id=post_);
    END;
$get_post_thread$ LANGUAGE plpgsql;

CREATE TABLE Post (
    author CITEXT REFERENCES Users (nickname) ON DELETE RESTRICT ON UPDATE RESTRICT NOT NULL,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    forum CITEXT REFERENCES Forum (slug) ON DELETE CASCADE ON UPDATE RESTRICT,
    id BIGSERIAL PRIMARY KEY,
    isEdited BOOLEAN NOT NULL DEFAULT false,
    message TEXT NOT NULL DEFAULT '',
    parent BIGINT REFERENCES Post (id) ON DELETE CASCADE ON UPDATE RESTRICT
        CONSTRAINT par CHECK (get_thread_by_post(parent)=thread),
    thread INTEGER REFERENCES Thread (id) ON DELETE CASCADE ON UPDATE RESTRICT,
    path bigint[] not null
);


CREATE TABLE vote (
    thread BIGINT REFERENCES Thread (id) ON DELETE CASCADE ON UPDATE CASCADE,
    author CITEXT REFERENCES Users (nickname) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL,
    vote SMALLINT CONSTRAINT check_vote CHECK (vote>=-1 AND vote <=1 ) DEFAULT 0,
    UNIQUE (thread, author)
);

CREATE OR REPLACE FUNCTION get_thread_by_post(post_ BIGINT) RETURNS INTEGER AS $get_post_thread$
    BEGIN
        RETURN (SELECT thread FROM Post WHERE id=post_);
    END;
$get_post_thread$ LANGUAGE plpgsql;

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

-- Триггер на Forum
CREATE OR REPLACE FUNCTION forum_user() RETURNS trigger AS $forum_user$
    BEGIN
        NEW.userNick = (SELECT nickname FROM Users WHERE lower(nickname)=lower(NEW.usernick));
        RETURN NEW;
    END
$forum_user$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS forum_user ON Forum;
CREATE TRIGGER forum_user BEFORE INSERT ON Forum
    FOR EACH ROW EXECUTE PROCEDURE forum_user();

-- Триггер на Post-ы. Отвечает за счетчики в Forum и за read-only данные Post
CREATE OR REPLACE FUNCTION update_forum_posts() RETURNS trigger AS $update_forum_posts$
    BEGIN
        IF TG_OP='INSERT' OR TG_OP='UPDATE' THEN
            NEW.forum = (SELECT forum FROM Thread WHERE Thread.id=NEW.thread);
        end if;
        IF TG_OP='INSERT' THEN
            UPDATE Forum SET posts=posts+1 WHERE slug=NEW.forum;
--             IF NEW.parent=0 THEN
--                 NEW.parent=NEW.id;
--             END IF;
            if NEW.forum IS NULL THEN
                RAISE NOTICE 'INSERT INTO Post';
--                 NEW.forum = (SELECT slug FROM Thread WHERE id=NEW.thread);
                NEW.forum = 'test_forum';
            end if;
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
            if NEW.message!=OLD.message OR NEW.parent!=OLD.parent then
                NEW.isEdited=TRUE;
            end if;
            RETURN NEW;
        END IF;

    END
$update_forum_posts$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS update_forum_posts ON Post;
CREATE TRIGGER update_forum_posts BEFORE UPDATE OR INSERT OR DELETE ON Post
    FOR EACH ROW EXECUTE PROCEDURE update_forum_posts();

CREATE OR REPLACE FUNCTION forum_user() RETURNS trigger AS $forum_user$
    BEGIN
        NEW.userNick = (SELECT nickname FROM Users WHERE lower(nickname)=lower(NEW.usernick));
        RETURN NEW;
    END
$forum_user$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS forum_user ON Forum;
CREATE TRIGGER forum_user BEFORE INSERT ON Forum
    FOR EACH ROW EXECUTE PROCEDURE forum_user();

CREATE OR REPLACE FUNCTION post_path() RETURNS TRIGGER AS
$post_path$
BEGIN
    if TG_OP='INSERT' then
        NEW.path = (SELECT path FROM Post WHERE id = NEW.parent) || NEW.id;
        RETURN NEW;
    end if;
END;
$post_path$ LANGUAGE plpgsql;

CREATE TRIGGER post_path BEFORE INSERT ON Post
    FOR EACH ROW EXECUTE PROCEDURE post_path();


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
            UPDATE Thread SET votes=votes+NEW.vote WHERE id=NEW.thread;
            RETURN NEW;
        ELSIF TG_OP='UPDATE' THEN
            UPDATE Thread SET votes=votes+(NEW.vote-OLD.vote) WHERE id=NEW.thread;
            RETURN NEW;
        ELSE
            RAISE EXCEPTION 'Invalid call update_thread_vote_counter()';
        end if;
    END
$update_thread_vote_counter$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS update_thread_vote ON Vote;
CREATE TRIGGER update_thread_vote AFTER INSERT OR UPDATE ON Vote
    FOR EACH ROW EXECUTE PROCEDURE update_thread_vote_counter();


CREATE OR REPLACE FUNCTION vote_thread(thread_ INTEGER, author_ citext, voice_ INTEGER) RETURNS void AS $update_thread_vote_counter$
    DECLARE
        currentVote SMALLINT;
    BEGIN
        currentVote := (SELECT vote FROM Vote WHERE thread=thread_ AND lower(author)=lower(author_));

        if currentVote IS NULL THEN
            INSERT INTO Vote (thread, author, vote) VALUES (thread_, author_, voice_);
        else
--             if currentVote*voice_<0 then --если нужно поменять запись
--                 UPDATE Vote SET vote=voice_ WHERE lower(thread)=lower(thread_) AND lower(author)=lower(author_);
--             end if;
            UPDATE Vote SET vote=voice_ WHERE thread=thread_ AND lower(author)=lower(author_);
--         IF voice_>0 THEN
--             INSERT INTO Vote (thread, author) VALUES (thread_, author_);
--         ELSE
--             DELETE FROM Vote WHERE lower(thread)=lower(thread_) AND lower(author)=lower(author_);
--         end if;
        end if;
    END
$update_thread_vote_counter$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION get_thread_id_by_slug(slugArg citext) RETURNS INTEGER AS $get_thread_id_by_slug$
    BEGIN
        RETURN (SELECT id FROM Thread WHERE lower(slug)=lower(slugArg));
    END
$get_thread_id_by_slug$ LANGUAGE plpgsql;

CREATE INDEX IF NOT EXISTS post_path_id ON Post (id, (path[1]));
CREATE INDEX IF NOT EXISTS post_path ON Post (path);
CREATE INDEX IF NOT EXISTS post_path_1 ON Post ((path[1]));
CREATE INDEX IF NOT EXISTS post_thread_id ON Post (thread, id);
CREATE INDEX IF NOT EXISTS post_thread ON Post (thread);
CREATE INDEX IF NOT EXISTS post_thread_path_id ON Post (thread, path, id);
CREATE INDEX IF NOT EXISTS post_thread_id_path_parent ON Post (thread, id, (path[1]), parent);
CREATE INDEX IF NOT EXISTS post_author_forum ON Post (author, forum);
CREATE INDEX IF NOT EXISTS idx_sth ON Post (lower(author));
CREATE INDEX IF not exists thread_slug ON Thread (slug);
CREATE INDEX IF NOT EXISTS thread_forum_created ON Thread (forum, created);
CREATE INDEX IF not exists thread_author_forum ON Thread (author, forum);
CREATE INDEX IF NOT EXISTS thread_author ON Thread (lower(author));
CREATE INDEX IF NOT EXISTS vote_nickname_thread ON Vote (author, thread);


INSERT INTO Users (email, fullname, nickname, about) VALUES ('ivanov.vanya@mail.ry', 'Ian', 'tamerlanchik', 'About me');
INSERT INTO Forum (slug, title, userNick) VALUES ('test_forum', 'Hello, world', 'tamerlanchik');
INSERT INTO Thread (author, forum, message, slug, title) VALUES ('tamerlanchik', 'test_forum', 'Hello fucking world!', 'test_thread', 'Fucking world!'), ('tamerlanchik', 'test_forum', 'Die', 'next_thread', 'DieDie');
INSERT INTO Post (author, forum, message, thread) VALUES ('tamerlanchik', 'test_forum', 'Hello', 1);