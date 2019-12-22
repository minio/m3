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


CREATE TABLE tasks
(
    id               SERIAL                                                    NOT NULL
        CONSTRAINT tasks_pk
            PRIMARY KEY,
    name             VARCHAR(256),
    status           VARCHAR                  DEFAULT 'new'::CHARACTER VARYING NOT NULL,
    data             JSONB,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now(),
    sys_updated_date TIMESTAMP WITH TIME ZONE,
    scheduled_time   TIMESTAMP WITH TIME ZONE
);

COMMENT ON COLUMN tasks.name IS 'name of the task that is scheduled';

COMMENT ON COLUMN tasks.data IS 'data to pass to the task';

COMMENT ON COLUMN tasks.scheduled_time IS 'time when the task was scheduled';

CREATE INDEX tasks_status_index
    ON tasks (status);

