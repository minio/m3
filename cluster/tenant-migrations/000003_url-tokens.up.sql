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

CREATE TABLE url_tokens
(
    id               UUID
        CONSTRAINT url_tokens_pk
            PRIMARY KEY,
    user_id          UUID                                   NOT NULL
        CONSTRAINT url_tokens_users_id_fk
            REFERENCES users
            ON DELETE CASCADE,
    expiration       TIMESTAMPTZ,
    used_for         VARCHAR(256),
    consumed         BOOL                     DEFAULT FALSE,
    sys_created_by   VARCHAR(256)                           NOT NULL,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

COMMENT ON TABLE url_tokens IS 'Holds tokens and their validity for multifple functions such as password reset or tenant invite';

COMMENT ON COLUMN url_tokens.user_id IS 'User this token is associated with';

COMMENT ON COLUMN url_tokens.expiration IS 'When does this token expires';

COMMENT ON COLUMN url_tokens.used_for IS 'describes the function this token is intenteded for (i.e. password-reset, signup-link)';

COMMENT ON COLUMN url_tokens.consumed IS 'whether or not the token has been already used';

