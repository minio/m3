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

--  Table to store sessions of a admins
CREATE TYPE ADMIN_STATUS_TYPE AS ENUM ('valid', 'invalid');

CREATE TABLE admin_sessions
(
    id                 VARCHAR(256)                           NOT NULL
        CONSTRAINT admin_sessions_pk
            PRIMARY KEY,                                                -- admin session id as rand string
    admin_id           UUID                                   NOT NULL
        CONSTRAINT admin_sessions_admins_id_fk
            REFERENCES admins
            ON DELETE CASCADE,                                          -- admin id of the user who initiated the session
    refresh_token      VARCHAR(256)                           NOT NULL, -- token used to request a new session
    occurred_at        TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL, -- first timestamp of the session
    last_event         TIMESTAMP WITH TIME ZONE DEFAULT now(),          -- stores last event's timestamp within this session
    expires_at         TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL, -- session's expiration time
    refresh_expires_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL, -- refresh token's expiration time
    status             ADMIN_STATUS_TYPE                      NOT NULL  -- session's status
);

CREATE UNIQUE INDEX admin_sessions_refresh_token_index
    ON admin_sessions (refresh_token);


