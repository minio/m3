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

CREATE INDEX permissions_sys_created_date_index
    ON permissions (sys_created_date);

CREATE INDEX service_accounts_sys_created_date_index
    ON service_accounts (sys_created_date);

CREATE INDEX users_sys_created_date_index
    ON users (sys_created_date);

ALTER TABLE bucket_metrics
    ADD COLUMN sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL;

CREATE INDEX bucket_metrics_sys_created_date_index
    ON bucket_metrics (sys_created_date);