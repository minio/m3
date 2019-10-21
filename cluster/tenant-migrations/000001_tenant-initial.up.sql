create table users
(
    id uuid not null
        constraint users_pk
            primary key,
    full_name varchar(256),
    email varchar(256) not null,
    password varchar(256),
    is_admin boolean default false,
    from_idp boolean
);

alter table users owner to postgres;

create unique index users_email_uindex
    on users (email);

create index users_hashed_password_index
    on users (password);

create table service_accounts
(
    id uuid not null
        constraint service_accounts_pk
            primary key,
    name varchar(256) not null,
    created_by varchar(256) not null,
    created_date timestamp with time zone default now() not null,
    description text
);

alter table service_accounts owner to postgres;

create table token_groups
(
    id uuid not null
        constraint token_groups_pk
            primary key,
    name varchar(256) not null,
    description text,
    created_by varchar(256) not null,
    created_date timestamp with time zone default now() not null
);

alter table token_groups owner to postgres;

create table tokens
(
    access_key varchar(256) not null
        constraint tokens_pk
            primary key,
    user_id uuid
        constraint tokens_users_id_fk
            references users
            on delete cascade,
    service_account_id uuid
        constraint tokens_service_accounts_id_fk
            references service_accounts
            on delete cascade,
    created_by varchar(256) not null,
    created_date timestamp with time zone default now() not null,
    ui_token boolean default false,
    token_group_id uuid not null
        constraint tokens_token_groups_id_fk
            references token_groups
            on delete cascade
);

alter table tokens owner to postgres;

create table policies
(
    name varchar(256) not null,
    created_by varchar(256) not null,
    created_date timestamp with time zone default now() not null,
    id uuid not null
        constraint policies_pk
            primary key
);

alter table policies owner to postgres;

create table groups
(
    id uuid not null
        constraint groups_pk
            primary key,
    name varchar(256) not null,
    policy_id uuid not null
        constraint groups_policies_id_fk
            references policies
            on delete cascade,
    created_by varchar(256) not null,
    created_date timestamp with time zone default now() not null
);

alter table groups owner to postgres;

create table users_groups
(
    user_id uuid not null
        constraint users_groups_users_id_fk
            references users
            on delete cascade,
    group_id uuid not null
        constraint users_groups_groups_id_fk
            references groups
            on delete cascade,
    created_by varchar(256) not null,
    created_date timestamp with time zone default now() not null
);

alter table users_groups owner to postgres;

create table policy_statements
(
    id uuid not null
        constraint policy_statements_pk
            primary key,
    policy_id uuid
        constraint policy_statements_policies_id_fk
            references policies
            on delete cascade,
    created_by varchar(256) not null,
    created_date timestamp with time zone default now() not null,
    effect varchar(64) not null
);

alter table policy_statements owner to postgres;

create table policy_statement_actions
(
    id uuid not null
        constraint policy_statement_actions_pk
            primary key,
    statement_id uuid
        constraint policy_statement_actions_policy_statements_id_fk
            references policy_statements
            on delete cascade,
    action varchar(256) not null,
    created_by varchar(256) not null,
    created_date timestamp with time zone default now() not null
);

alter table policy_statement_actions owner to postgres;

create table policy_statement_resources
(
    id uuid not null
        constraint policy_statement_resources_pk
            primary key,
    statement_id uuid
        constraint policy_statement_resources_policy_statements_id_fk
            references policy_statements,
    resource varchar(512) not null,
    created_by varchar(256) not null,
    created_date timestamp with time zone default now() not null
);

alter table policy_statement_resources owner to postgres;

create table service_account_groups
(
    service_account_id uuid not null
        constraint service_account_groups_service_accounts_id_fk
            references service_accounts
            on delete cascade,
    group_id uuid not null
        constraint service_account_groups_groups_id_fk
            references groups
            on delete cascade,
    created_by varchar(256) not null,
    created_date timestamp with time zone default now() not null
);

alter table service_account_groups owner to postgres;

