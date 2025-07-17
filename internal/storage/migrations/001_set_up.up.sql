CREATE TABLE events (
    id       varchar not null,
    group_id varchar not null,
    data     bytea   not null,
    CONSTRAINT events_pk
        PRIMARY KEY (id));