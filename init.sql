CREATE EXTENSION IF NOT EXISTS citext;
DROP TABLE IF EXISTS ForumPosts;
DROP TABLE If EXISTS UsersInForum;
DROP TABLE If EXISTS ThreadVotes;
DROP TABLE IF EXISTS Vote;
DROP TABLE IF EXISTS Post;
DROP TABLE IF EXISTS Thread;
DROP TABLE IF EXISTS Forum;
DROP TABLE IF EXISTS Users;

CREATE UNLOGGED TABLE Users (
    about TEXT NOT NULL DEFAULT '',
    email CITEXT NOT NULL UNIQUE CONSTRAINT email_right CHECK(email ~ '^.*@[A-Za-z0-9\-_\.]*$'),
    fullname TEXT NOT NULL DEFAULT '',
    nickname CITEXT COLLATE "POSIX" PRIMARY KEY CONSTRAINT nick_right CHECK(nickname ~ '^[A-Za-z0-9_\.]*$')
);

-- _________Forum____________
CREATE UNLOGGED TABLE Forum (
    slug CITEXT PRIMARY KEY UNIQUE CONSTRAINT slug_correct CHECK(slug ~ '^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$'),
    threads INTEGER DEFAULT 0,
    title TEXT NOT NULL DEFAULT '',
    userNick CITEXT REFERENCES Users (nickname) ON DELETE RESTRICT ON UPDATE RESTRICT NOT NULL,
    posts INTEGER DEFAULT 0
);

create unlogged table ForumPosts (
    forum citext PRIMARY KEY,
    posts INTEGER DEFAULT 0
);


-- _______Thread__________
CREATE unlogged TABLE Thread (
    author CITEXT REFERENCES Users (nickname) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    forum CITEXT REFERENCES Forum (slug) ON DELETE CASCADE ON UPDATE RESTRICT not null,
    id SERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    slug CITEXT UNIQUE CONSTRAINT slug_correct CHECK(slug ~ '^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$'),
    title TEXT NOT NULL,
    votes INTEGER NOT NULL DEFAULT 0
);

-- _______Post___________
CREATE OR REPLACE FUNCTION get_thread_by_post(post_ BIGINT) RETURNS INTEGER AS $$
    BEGIN
        RETURN (SELECT thread FROM Post WHERE id=post_);
    END;
$$ LANGUAGE plpgsql;

CREATE unlogged TABLE Post (
    author CITEXT REFERENCES Users (nickname) NOT NULL,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    forum CITEXT,
    id SERIAL PRIMARY KEY,
    isEdited BOOLEAN NOT NULL DEFAULT false,
    message TEXT NOT NULL DEFAULT '',
    parent INTEGER REFERENCES Post (id) ON DELETE CASCADE ON UPDATE RESTRICT
        CONSTRAINT par CHECK (get_thread_by_post(parent)=thread),
    thread INTEGER,
    path INTEGER[] not null
);

-- ______Vote______
CREATE unlogged TABLE Vote (
    thread INTEGER,
    author CITEXT REFERENCES Users (nickname) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL,
    vote SMALLINT DEFAULT 0,
    UNIQUE (thread, author)
);

create unlogged table UsersInForum (
    nickname CITEXT COLLATE "POSIX",
    forum citext
);

------------Триггеры-------------------------

-- ____THREAD______

-- Добавляет создателя Thread в табличку ForumUsers
create or replace function users_forum() returns trigger as $$
    begin
        if NEW.forum IS NOT NULL then
            INSERT INTO usersinforum(forum, nickname) VALUES (NEW.forum, new.author) on conflict do nothing;
        end if;
        RETURN new;
    end;
$$ language plpgsql;
drop trigger if exists users_forum on Thread;
create trigger users_forum after insert on Thread
    for each row  execute procedure users_forum();

-- Отвечает за счетчик в Forums и read-only поля Thread
CREATE OR REPLACE FUNCTION update_forum_threads() RETURNS trigger AS $update_forum_threads$
    BEGIN
        IF TG_OP='INSERT' THEN
            UPDATE Forum SET threads=threads+1 WHERE slug=NEW.forum;
            RETURN NEW;
        ELSIF TG_OP='DELETE' OR TG_OP='TRUNCATE' THEN
            UPDATE Forum SET threads=threads-1 WHERE slug=OLD.forum;
            RETURN OLD;
        ELSIF TG_OP='UPDATE' THEN
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
        RETURN NEW;
    END
$update_forum_threads$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS update_forum_threads ON Thread;
CREATE TRIGGER update_forum_threads AFTER UPDATE OR INSERT OR DELETE ON Thread
    FOR EACH ROW EXECUTE PROCEDURE update_forum_threads();

-- _____USERS_____

-- Read-only поля Users
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

-- _____FORUM_____

-- Заменяет переданное (возможно, со сбитым регистром)
--        имя пользователя соответствующим в БД
CREATE OR REPLACE FUNCTION forum_user() RETURNS trigger AS $forum_user$
    BEGIN
        NEW.userNick = (SELECT nickname FROM Users WHERE lower(nickname)=lower(NEW.usernick));
        RETURN NEW;
    END
$forum_user$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS forum_user ON Forum;
CREATE TRIGGER forum_user BEFORE INSERT ON Forum
    FOR EACH ROW EXECUTE PROCEDURE forum_user();

-- ______POST_______

-- Отвечает за счетчики в Forum и за read-only данные Post
CREATE OR REPLACE FUNCTION update_forum_posts() RETURNS trigger AS $update_forum_posts$
    BEGIN
        IF TG_OP='DELETE' OR TG_OP='TRUNCATE' THEN
            UPDATE ForumPosts SET posts=posts-1 WHERE forum=OLD.forum;
            RETURN OLD;
        ELSIF TG_OP='UPDATE' THEN
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
        RETURN NEW;
    END
$update_forum_posts$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS update_forum_posts ON Post;
CREATE TRIGGER update_forum_posts BEFORE UPDATE OR DELETE ON Post
    FOR EACH ROW EXECUTE PROCEDURE update_forum_posts();

-- Выставляет path
CREATE OR REPLACE FUNCTION post_path() RETURNS TRIGGER AS
$$
BEGIN
    NEW.path = (SELECT path FROM Post WHERE id = NEW.parent) || NEW.id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER post_path BEFORE INSERT ON Post
    FOR EACH ROW EXECUTE PROCEDURE post_path();

-- ____VOTE_____
-- Отвечает за счетчики votes у Thread.
CREATE OR REPLACE FUNCTION update_thread_vote_counter() RETURNS trigger AS $$
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
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS update_thread_vote ON Vote;
CREATE TRIGGER update_thread_vote AFTER INSERT OR UPDATE ON Vote
    FOR EACH ROW EXECUTE PROCEDURE update_thread_vote_counter();

-- Индексы

CREATE UNIQUE INDEX users_nickname_index on Users (LOWER(nickname));
CREATE UNIQUE INDEX forum_slug_index on Forum (LOWER(slug));

create index IF NOT EXISTS post__thread ON Post(thread);
create index IF NOT EXISTS post__id_thread ON post(id, thread);
create index IF NOT EXISTS post__path__first ON Post((path[1]));
create index IF NOT EXISTS post_forum_author ON post(forum, author);
create index post_parent_thread_path_id ON Post(thread, (path[1]), id) WHERE parent IS NUll;
CREATE INDEX IF NOT EXISTS idx_sth ON Post (lower(author));

CREATE UNIQUE INDEX thread_slug_index on Thread (LOWER(slug));
CREATE INDEX IF NOT EXISTS thread_author ON Thread (lower(author));
create index IF NOT EXISTS thread_forum ON thread(forum);
create index IF NOT EXISTS thread_forum_created on Thread(lower(forum), created);

create index IF NOT EXISTS vote_coverage On Vote(thread, lower(author), vote);

create unique index forum_users_idx ON UsersInForum(forum, nickname);
cluster UsersInForum USING forum_users_idx;
