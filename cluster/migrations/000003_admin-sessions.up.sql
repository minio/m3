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
CREATE TYPE admin_status_type AS ENUM ('valid', 'invalid');

CREATE TABLE admin_sessions
(
    id            varchar(256)                           not null
        constraint admin_sessions_pk
            primary key,                                           -- admin session id as rand string
    admin_id      uuid                                   not null
        constraint admin_sessions_admins_id_fk
            references admins
            on delete cascade,                                     -- admin id of the user who initiated the session
    refresh_token varchar(256)                           not null, -- token used to request a new session
    occurred_at   timestamp with time zone default now() not null, -- first timestamp of the session
    last_event    timestamp with time zone default now(),          -- stores last event's timestamp within this session
    expires_at    timestamp with time zone default now() not null, -- session's expiration time
    refresh_expires_at    timestamp with time zone default now() not null, -- refresh token's expiration time
    status        admin_status_type                      not null  -- session's status
);

create unique index admin_sessions_refresh_token_index
    on admin_sessions (refresh_token);


