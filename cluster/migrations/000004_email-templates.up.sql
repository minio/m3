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

CREATE TABLE email_templates
(
    name       VARCHAR(256) NOT NULL
        CONSTRAINT email_templates_pk
            PRIMARY KEY,
    template TEXT         NOT NULL
);

COMMENT ON TABLE email_templates IS 'Table to store the email templates';

COMMENT ON COLUMN email_templates.name IS 'Identifies the template by it''s key name, for example `signup` will have the template for the signup email';


