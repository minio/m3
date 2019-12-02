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

CREATE UNIQUE INDEX nodes_name_uindex
    ON nodes (name);

CREATE UNIQUE INDEX storage_clusters_name_uindex
    ON storage_clusters (name);

ALTER TABLE storage_groups
    ALTER COLUMN name SET NOT NULL;

CREATE UNIQUE INDEX storage_groups_name_uindex
    ON storage_groups (name);

DROP TABLE IF EXISTS storage_clusters_groups;

ALTER TABLE storage_groups
    ADD storage_cluster_id UUID NOT NULL;

ALTER TABLE storage_groups
    ADD CONSTRAINT storage_groups_storage_clusters_id_fk
        FOREIGN KEY (storage_cluster_id) REFERENCES storage_clusters;

ALTER TABLE storage_cluster_nodes
    DROP COLUMN k8s_label;

ALTER TABLE storage_cluster_nodes
    ADD num INT NOT NULL;

ALTER TABLE node_volumes
    ADD num INT NOT NULL;



