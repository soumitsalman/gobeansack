
INSTALL vss;
LOAD vss;

CREATE TABLE IF NOT EXISTS beans (
    url VARCHAR PRIMARY KEY,
    kind VARCHAR NOT NULL,
    title VARCHAR,
    title_length INTEGER DEFAULT 0,
    content TEXT,
    content_length INTEGER DEFAULT 0,
    summary TEXT,
    summary_length INTEGER DEFAULT 0,
    author VARCHAR,
    source VARCHAR,
    created TIMESTAMP,
    collected TIMESTAMP
);

CREATE TABLE IF NOT EXISTS generated_beans (
    url VARCHAR PRIMARY KEY,
    intro TEXT,
    analysis TEXT[],
    insights TEXT[],
    verdict TEXT,
    predictions TEXT[]
);
CREATE TABLE IF NOT EXISTS bean_embeddings (
    url VARCHAR PRIMARY KEY,
    embedding FLOAT[%d] NOT NULL
);
CREATE TABLE IF NOT EXISTS bean_clusters (
    url VARCHAR NOT NULL,
    tag VARCHAR NOT NULL
);
CREATE TABLE IF NOT EXISTS bean_categories (
    url VARCHAR NOT NULL,
    tag VARCHAR NOT NULL
);
CREATE TABLE IF NOT EXISTS bean_sentiments (
    url VARCHAR NOT NULL,
    tag VARCHAR NOT NULL
);
CREATE TABLE IF NOT EXISTS bean_gists (
    url VARCHAR NOT NULL,
    tag TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS bean_regions (
    url VARCHAR NOT NULL,
    tag VARCHAR NOT NULL
);
CREATE TABLE IF NOT EXISTS bean_entities (
    url VARCHAR NOT NULL,
    tag VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS chatters (
    chatter_url VARCHAR NOT NULL,
    bean_url VARCHAR NOT NULL,
    collected TIMESTAMP NOT NULL,
    source VARCHAR NOT NULL,
    forum VARCHAR,
    likes INTEGER DEFAULT 0,
    comments INTEGER DEFAULT 0,
    subscribers INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS sources (
    name VARCHAR,
    description TEXT,
    base_url VARCHAR PRIMARY KEY,
    domain_name VARCHAR,
    favicon VARCHAR,
    rss_feed VARCHAR
);

CREATE TABLE IF NOT EXISTS categories AS
SELECT 
    _id as tag,
    embedding::FLOAT[%d] as embedding
FROM read_parquet('%s');

CREATE TABLE IF NOT EXISTS sentiments AS
SELECT 
    _id as tag,
    embedding::FLOAT[%d] as embedding
FROM read_parquet('%s');

