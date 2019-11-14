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

create table admin_tokens
(
    id               uuid
        constraint admin_tokens_pk
            primary key,
    admin_id         uuid                                   not null
        constraint admin_tokens_admin_id_fk
            references admins
            on delete cascade,
    expiration       timestamptz,
    used_for         varchar(256),
    consumed         bool                     default false,
    sys_created_by   varchar(256)                           not null,
    sys_created_date timestamp with time zone default now() not null
);

comment on table admin_tokens is 'Holds tokens and their validity for multifple functions such as password reset or tenant invite';

comment on column admin_tokens.admin_id is 'Admin this token is associated with';

comment on column admin_tokens.expiration is 'When does this token expires';

comment on column admin_tokens.used_for is 'describes the function this token is intenteded for (i.e. password-reset, signup-link)';

comment on column admin_tokens.consumed is 'whether or not the token has been already used';

