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

ALTER TABLE bucket_metrics
	DROP COLUMN buckets_sizes;

ALTER TABLE bucket_metrics
	DROP COLUMN total_cost;

ALTER TABLE bucket_metrics
    ADD COLUMN bucket_name VARCHAR(256);

ALTER TABLE bucket_metrics
    ADD COLUMN bucket_size NUMERIC(18,0); -- size in bytes, up to 100 petabyte

CREATE INDEX bucket_metrics_last_update_index
    ON bucket_metrics (last_update);