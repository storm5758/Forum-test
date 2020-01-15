CREATE EXTENSION IF NOT EXISTS citext;

-- ######################################################
-- USERS
-- ######################################################
DROP TABLE IF EXISTS users CASCADE;
CREATE UNLOGGED TABLE users(
    about TEXT,
    email CITEXT NOT NULL UNIQUE CONSTRAINT email_right CHECK(email ~ '^.*@[A-Za-z0-9\-_\.]*$'),
    fullname TEXT NOT NULL,
    nickname CITEXT PRIMARY KEY COLLATE "POSIX" CONSTRAINT nick_right CHECK(nickname ~ '^[A-Za-z0-9_\.]*$')
);
-- indexes
CREATE INDEX IF NOT EXISTS idx_users_nick ON users (LOWER(nickname));
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- ######################################################
-- Forums
-- ######################################################
DROP TABLE IF EXISTS forums CASCADE;
CREATE UNLOGGED TABLE forums (
    slug CITEXT PRIMARY KEY UNIQUE CONSTRAINT slug_correct CHECK(slug ~ '^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$'),
    threads INTEGER DEFAULT 0,
    posts INTEGER DEFAULT 0,
    user_nick CITEXT REFERENCES Users (nickname) ON DELETE RESTRICT ON UPDATE RESTRICT NOT NULL,
    title TEXT NOT NULL
);

-- indexes
CREATE INDEX IF NOT EXISTS idx_forum_slug ON forums (LOWER(slug));

-- ######################################################
-- USERS IN FORUM
-- ######################################################
DROP TABLE IF EXISTS UsersInForum CASCADE;
CREATE UNLOGGED TABLE UsersInForum (
    nickname CITEXT COLLATE "POSIX",
    forum CITEXT
);
-- indexes
CREATE UNIQUE INDEX forum_users_idx ON UsersInForum(forum, nickname);

-- ######################################################
-- Threads
-- ######################################################
DROP TABLE IF EXISTS threads CASCADE;
CREATE UNLOGGED TABLE threads (
    id SERIAL PRIMARY KEY,
    slug CITEXT UNIQUE CONSTRAINT slug_correct CHECK(slug ~ '^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$'),
    forum CITEXT REFERENCES forums (slug) ON DELETE CASCADE ON UPDATE RESTRICT NOT NULL,
    author CITEXT REFERENCES users (nickname) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL,
    created TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    message TEXT NOT NULL,
    title TEXT NOT NULL,
    votes INTEGER NOT NULL DEFAULT 0
);
-- indexes
CREATE INDEX IF NOT EXISTS thread_id ON threads (LOWER(slug), id);
CREATE UNIQUE INDEX thread_slug_index ON threads (LOWER(slug));
CREATE INDEX IF NOT EXISTS thread_author ON threads (LOWER(author));
CREATE INDEX IF NOT EXISTS thread_forum ON threads (forum);
CREATE INDEX IF NOT EXISTS thread_forum_created on threads(LOWER(forum), created);

-- ######################################################
-- VOTES
-- ######################################################
DROP TABLE IF EXISTS votes CASCADE;
CREATE UNLOGGED TABLE votes (
    thread BIGINT REFERENCES threads (id) ON DELETE CASCADE ON UPDATE CASCADE,
    author CITEXT REFERENCES users (nickname) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL,
    vote SMALLINT CONSTRAINT check_vote CHECK (vote = -1 OR vote = 1 ) DEFAULT 0,
    UNIQUE (thread, author)
);
-- indexes
CREATE INDEX IF NOT EXISTS vote_coverage ON votes(thread, lower(author), vote);

-- ######################################################
-- POSTS
-- ######################################################
CREATE OR REPLACE FUNCTION get_thread_by_post(post_ BIGINT) RETURNS INTEGER AS $$
    BEGIN
        RETURN (SELECT thread FROM posts WHERE id = post_);
    END;
$$ LANGUAGE plpgsql;

DROP TABLE IF EXISTS posts CASCADE;
CREATE UNLOGGED TABLE posts (
    author CITEXT REFERENCES users (nickname) NOT NULL,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    forum CITEXT,
    id SERIAL PRIMARY KEY,
    isEdited BOOLEAN NOT NULL DEFAULT false,
    message TEXT NOT NULL DEFAULT '',
    parent INTEGER REFERENCES posts (id) ON DELETE CASCADE ON UPDATE RESTRICT
        CONSTRAINT par CHECK (get_thread_by_post(parent)=thread),
    thread INTEGER,
    path INTEGER[] not null
);
-- indexes
CREATE INDEX IF NOT EXISTS post__thread ON posts(thread);
CREATE INDEX IF NOT EXISTS post__id_thread ON posts(id, thread);
CREATE INDEX IF NOT EXISTS post__path__first ON posts((path[1]));
CREATE INDEX IF NOT EXISTS post_forum_author ON posts(forum, author);
CREATE INDEX post_parent_thread_path_id ON posts(thread, (path[1]), id) WHERE parent IS NUll;
CREATE INDEX IF NOT EXISTS idx_sth ON posts (LOWER(author));

-- ######################################################
-- FORUM-POSTS
-- ######################################################
DROP TABLE IF EXISTS ForumPosts CASCADE;
CREATE UNLOGGED TABLE ForumPosts (
    forum citext PRIMARY KEY,
    posts INTEGER DEFAULT 0
);

-- functions
-- AUTO ForumPosts
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
-- triggers
DROP TRIGGER IF EXISTS update_forum_posts ON posts;
CREATE TRIGGER update_forum_posts BEFORE UPDATE OR DELETE ON posts
    FOR EACH ROW EXECUTE PROCEDURE update_forum_posts();

-- functions
-- AUTO threads+- IN FORUMS
CREATE OR REPLACE FUNCTION update_forum_threads() RETURNS trigger AS $update_forum_threads$
    BEGIN
        IF TG_OP='INSERT' THEN
            UPDATE forums SET threads=threads+1 WHERE slug=NEW.forum;
            RETURN NEW;
        ELSIF TG_OP='DELETE' OR TG_OP='TRUNCATE' THEN
            UPDATE forums SET threads=threads-1 WHERE slug=OLD.forum;
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

DROP TRIGGER IF EXISTS update_forum_threads ON threads;
CREATE TRIGGER update_forum_threads AFTER UPDATE OR INSERT OR DELETE ON threads
    FOR EACH ROW EXECUTE PROCEDURE update_forum_threads();
-- functions
-- AUTO INSERT TO UsersInForum
CREATE OR REPLACE FUNCTION users_forum() RETURNS trigger AS $$
    BEGIN
        IF NEW.forum IS NOT NULL THEN
            INSERT INTO UsersInForum(forum, nickname) VALUES (NEW.forum, new.author) ON conflict do nothing;
        END IF;
        RETURN new;
    END;
$$ language plpgsql;
-- trigers
DROP trigger IF EXISTS users_forum ON threads;
CREATE trigger users_forum AFTER INSERT ON threads
    FOR EACH ROW EXECUTE PROCEDURE users_forum();

-- AUTO votes+- IN THREADS

CREATE OR REPLACE FUNCTION update_thread_vote_counter() RETURNS trigger AS $$
    BEGIN
        IF TG_OP='INSERT' THEN
            UPDATE threads SET votes=votes+NEW.vote WHERE id=NEW.thread;
            RETURN NEW;
        ELSIF TG_OP='UPDATE' THEN
            UPDATE threads SET votes=votes+(NEW.vote-OLD.vote) WHERE id=NEW.thread;
            RETURN NEW;
        ELSE
            RAISE EXCEPTION 'Invalid call update_thread_vote_counter()';
        END IF;
    END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_thread_vote ON votes;
CREATE TRIGGER update_thread_vote AFTER INSERT OR UPDATE ON votes
    FOR EACH ROW EXECUTE PROCEDURE update_thread_vote_counter();

--functions
-- AUTO path IN POSTS
CREATE OR REPLACE FUNCTION post_path() RETURNS TRIGGER AS
$$
BEGIN
    NEW.path = (SELECT path FROM posts WHERE id = NEW.parent) || NEW.id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- triggers
DROP TRIGGER IF EXISTS post_path ON posts;
CREATE TRIGGER post_path BEFORE INSERT ON posts
    FOR EACH ROW EXECUTE PROCEDURE post_path();

CREATE OR REPLACE FUNCTION forum_user() RETURNS trigger AS $forum_user$
    BEGIN
        NEW.user_nick = (SELECT nickname FROM users WHERE lower(nickname)=lower(NEW.user_nick));
        RETURN NEW;
    END
$forum_user$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS forum_user ON forums;
CREATE TRIGGER forum_user BEFORE INSERT ON forums
    FOR EACH ROW EXECUTE PROCEDURE forum_user();

cluster UsersInForum USING forum_users_idx;