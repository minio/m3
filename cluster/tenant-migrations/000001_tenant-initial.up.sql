create table users
(
    id           uuid                                   not null
        constraint users_pk
            primary key,
    full_name    varchar(256),
    email        varchar(256)                           not null,
    password     varchar(256),
    is_admin     boolean                  default false,
    from_idp     boolean,
    sys_created_date timestamp with time zone default now() not null
);


create unique index users_email_uindex
    on users (email);

create index users_password_index
    on users (password);

create table service_accounts
(
    id           uuid                                   not null
        constraint service_accounts_pk
            primary key,
    name         varchar(256)                           not null,
    description  text,
    sys_created_by   varchar(256)                           not null,
    sys_created_date timestamp with time zone default now() not null,
    sys_deleted  boolean                  default false
);

create index service_accounts_sys_deleted_index
    on service_accounts (sys_deleted);

create table credentials
(
    access_key         varchar(256)                           not null
        constraint credentials_pk
            primary key,
    user_id            uuid
        constraint credentials_users_id_fk
            references users
            on delete cascade,
    service_account_id uuid
        constraint credentials_service_accounts_id_fk
            references service_accounts
            on delete cascade,
    ui_credential      boolean                  default false,

    sys_created_by         varchar(256)                           not null,
    sys_created_date       timestamp with time zone default now() not null,
    sys_deleted        boolean                  default false
);

create index credentials_sys_deleted_index
    on credentials (sys_deleted);


create table permissions
(
    id           uuid                                   not null
        constraint permissions_pk
            primary key,
    effect       varchar(64)                            not null,
    sys_created_by   varchar(256)                           not null,
    sys_created_date timestamp with time zone default now() not null
);


create table permissions_resources
(
    id            uuid                                   not null
        constraint permissions_resources_pk
            primary key,
    permission_id uuid
        constraint permissions_resources_permissions_id_fk
            references permissions,
    resource      varchar(512)                           not null,
    sys_created_by    varchar(256)                           not null,
    sys_created_date  timestamp with time zone default now() not null
);


create table service_accounts_permissions
(
    service_account_id uuid                                   not null
        constraint service_accounts_permissions_service_accounts_id_fk
            references service_accounts
            on delete cascade,
    permission_id      uuid                                   not null
        constraint service_accounts_permissions_permissions_id_fk
            references permissions
            on delete cascade,
    sys_created_by         varchar(256)                           not null,
    sys_created_date       timestamp with time zone default now() not null
);


create table actions
(
    id          uuid not null
        constraint actions_pk
            primary key,
    name        varchar(256),
    description text
);


create table permissions_actions
(
    permission_id uuid
        constraint permissions_actions_permissions_id_fk
            references permissions
            on delete cascade,
    action_id     uuid                                   not null
        constraint permissions_actions_actions_id_fk
            references actions,
    sys_created_by    varchar(256)                           not null,
    sys_created_date  timestamp with time zone default now() not null
);


create table api_logs
(
    id           serial                                 not null
        constraint api_logs_pk
            primary key,
    api          varchar(256)                           not null,
    payload      text,
    sys_created_date timestamp with time zone default now() not null,
    session_id   varchar(256),
    user_email   varchar(256)
);


create index api_logs_api_index
    on api_logs (api);

create index api_logs_session_id_index
    on api_logs (session_id);

create index api_logs_user_email_index
    on api_logs (user_email);


