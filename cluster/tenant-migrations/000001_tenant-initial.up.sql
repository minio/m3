-- This file is part of MinIO Kubernetes Cloud
-- Copyright (c) 2019 MinIO, Inc.
--
-- This program is free software: you can redistribute it and/or modify
-- it under the terms of the GNU Affero General Public License as published by
-- the Free Software Foundation, either version 3 of the License, or
-- (at your option) any later version.
--
-- This program is distributed in the hope that it will be useful,
-- but WITHOUT ANY WARRANTY; without even the implied warranty of
-- MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
-- GNU Affero General Public License for more details.
--
-- You should have received a copy of the GNU Affero General Public License
-- along with this program.  If not, see <http://www.gnu.org/licenses/>.

CREATE TABLE users
(
    id                  UUID                                   NOT NULL
        CONSTRAINT users_pk
            PRIMARY KEY,
    full_name           VARCHAR(256),
    email               VARCHAR(256)                           NOT NULL,
    password            VARCHAR(256),
    accepted_invitation BOOLEAN                  DEFAULT FALSE,
    sys_created_date    TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);


CREATE UNIQUE INDEX users_email_uindex
    ON users (email);

CREATE INDEX users_password_index
    ON users (password);

CREATE TABLE service_accounts
(
    id               UUID                                   NOT NULL
        CONSTRAINT service_accounts_pk
            PRIMARY KEY,
    name             VARCHAR(256)                           NOT NULL,
    slug             VARCHAR(256)                           NOT NULL,
    description      TEXT,
    sys_created_by   VARCHAR(256)                           NOT NULL,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
    sys_deleted      TIMESTAMP WITH TIME ZONE
);

CREATE INDEX service_accounts_sys_deleted_index
    ON service_accounts (sys_deleted);

CREATE UNIQUE INDEX service_accounts_slug_index
    ON service_accounts (slug);


CREATE TABLE credentials
(
    access_key         VARCHAR(256)                           NOT NULL
        CONSTRAINT credentials_pk
            PRIMARY KEY,
    user_id            UUID
        CONSTRAINT credentials_users_id_fk
            REFERENCES users
            ON DELETE CASCADE,
    service_account_id UUID
        CONSTRAINT credentials_service_accounts_id_fk
            REFERENCES service_accounts
            ON DELETE CASCADE,
    ui_credential      BOOLEAN                  DEFAULT FALSE,

    sys_created_by     VARCHAR(256)                           NOT NULL,
    sys_created_date   TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
    sys_deleted        TIMESTAMP WITH TIME ZONE
);

CREATE INDEX credentials_sys_deleted_index
    ON credentials (sys_deleted);


CREATE TABLE permissions
(
    id               UUID                                   NOT NULL
        CONSTRAINT permissions_pk
            PRIMARY KEY,
    name             VARCHAR(512),
    slug             VARCHAR(512)                           NOT NULL,
    description      TEXT,
    effect           VARCHAR(64)                            NOT NULL,
    sys_created_by   VARCHAR(256)                           NOT NULL,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX permissions_slug_index
    ON permissions (slug);


CREATE TABLE permissions_resources
(
    id               UUID                                   NOT NULL
        CONSTRAINT permissions_resources_pk
            PRIMARY KEY,
    permission_id    UUID
        CONSTRAINT permissions_resources_permissions_id_fk
            REFERENCES permissions
            ON DELETE CASCADE,
    bucket_name      VARCHAR(64)                            NOT NULL,
    path             VARCHAR(512)                           NOT NULL,
    sys_created_by   VARCHAR(256)                           NOT NULL,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

CREATE TABLE permissions_actions
(
    id               UUID                                   NOT NULL
        CONSTRAINT permissions_actions_pk
            PRIMARY KEY,
    permission_id    UUID
        CONSTRAINT permissions_actions_permissions_id_fk
            REFERENCES permissions
            ON DELETE CASCADE,
    action           VARCHAR(256)                           NOT NULL,
    sys_created_by   VARCHAR(256)                           NOT NULL,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);


CREATE TABLE service_accounts_permissions
(
    service_account_id UUID                                   NOT NULL
        CONSTRAINT service_accounts_permissions_service_accounts_id_fk
            REFERENCES service_accounts
            ON DELETE CASCADE,
    permission_id      UUID                                   NOT NULL
        CONSTRAINT service_accounts_permissions_permissions_id_fk
            REFERENCES permissions
            ON DELETE CASCADE,
    sys_created_by     VARCHAR(256)                           NOT NULL,
    sys_created_date   TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX service_accounts_permissions_service_account_id_permission_id_u
    ON service_accounts_permissions (service_account_id, permission_id);

CREATE TABLE api_logs
(
    id               SERIAL                                 NOT NULL
        CONSTRAINT api_logs_pk
            PRIMARY KEY,
    api              VARCHAR(256)                           NOT NULL,
    payload          TEXT,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
    session_id       VARCHAR(256),
    user_email       VARCHAR(256)
);


CREATE INDEX api_logs_api_index
    ON api_logs (api);

CREATE INDEX api_logs_session_id_index
    ON api_logs (session_id);

CREATE INDEX api_logs_user_email_index
    ON api_logs (user_email);


