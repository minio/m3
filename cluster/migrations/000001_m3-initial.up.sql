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

CREATE TABLE admins
(
    id               UUID
        CONSTRAINT admins_pk
            PRIMARY KEY,
    name             VARCHAR(256),
    email            VARCHAR(256)              NOT NULL,
    password         VARCHAR(256),
    sys_created_by   VARCHAR(256)              NOT NULL,
    sys_created_date TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX admins_email_uindex
    ON admins (email);

CREATE TABLE tenants
(
    id               UUID                                   NOT NULL
        CONSTRAINT tenants_pk
            PRIMARY KEY,
    name             VARCHAR(256)                           NOT NULL,
    short_name       VARCHAR(256)                           NOT NULL,
    sys_created_by   VARCHAR(256)                           NOT NULL,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);


CREATE TABLE storage_groups
(
    id               UUID                                   NOT NULL
        CONSTRAINT storage_groups_pk
            PRIMARY KEY,
    name             VARCHAR(256),
    num              SERIAL                                 NOT NULL,
    sys_created_by   VARCHAR(256)                           NOT NULL,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);


CREATE TABLE tenants_storage_groups
(
    tenant_id        UUID                                   NOT NULL
        CONSTRAINT tenants_storage_groups_tenants_id_fk
            REFERENCES tenants,
    storage_group_id UUID                                   NOT NULL
        CONSTRAINT tenants_storage_groups_storage_groups_id_fk
            REFERENCES storage_groups,
    port             INTEGER                                NOT NULL,
    service_name     VARCHAR(64)                            NOT NULL,
    sys_created_by   VARCHAR(256)                           NOT NULL,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);


CREATE TABLE nodes
(
    id               UUID                                   NOT NULL
        CONSTRAINT nodes_pk
            PRIMARY KEY,
    name             VARCHAR(256),
    k8s_label        VARCHAR(256),
    sys_created_by   VARCHAR(256)                           NOT NULL,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);


CREATE TABLE storage_clusters
(
    id               UUID                                   NOT NULL
        CONSTRAINT storage_clusters_pk
            PRIMARY KEY,
    name             VARCHAR(256),
    sys_created_by   VARCHAR(256)                           NOT NULL,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);


CREATE TABLE storage_clusters_groups
(
    storage_cluster_id UUID                                   NOT NULL
        CONSTRAINT storage_clusters_groups_storage_clusters_id_fk
            REFERENCES storage_clusters,
    storage_group_id   UUID                                   NOT NULL
        CONSTRAINT storage_clusters_groups_storage_groups_id_fk
            REFERENCES storage_groups,
    sys_created_by     VARCHAR(256)                           NOT NULL,
    sys_created_date   TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);


CREATE TABLE storage_cluster_nodes
(
    storage_cluster_id UUID                                   NOT NULL
        CONSTRAINT storage_cluster_nodes_storage_clusters_id_fk
            REFERENCES storage_clusters,
    node_id            UUID                                   NOT NULL
        CONSTRAINT storage_cluster_nodes_nodes_id_fk
            REFERENCES nodes,
    k8s_label          VARCHAR(256),
    sys_created_by     VARCHAR(256)                           NOT NULL,
    sys_created_date   TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);


CREATE TABLE node_volumes
(
    id               UUID                                   NOT NULL
        CONSTRAINT node_volumes_pk
            PRIMARY KEY,
    node_id          UUID                                   NOT NULL
        CONSTRAINT node_volumes_nodes_id_fk
            REFERENCES nodes,
    mount_path       VARCHAR(256),
    sys_created_by   VARCHAR(256)                           NOT NULL,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

--  Table to store Disks attached to a node and their mount points

CREATE TABLE disks
(
    id               UUID                                   NOT NULL
        CONSTRAINT disks_pk
            PRIMARY KEY,
    node_id          UUID
        CONSTRAINT disks_nodes_id_fk
            REFERENCES nodes,
    mount_point      VARCHAR(512),
    capacity         BIGINT,
    sys_created_by   VARCHAR(256)                           NOT NULL,
    sys_created_date TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

COMMENT ON COLUMN disks.capacity IS 'Capacity in bytes';

--  Table to store sessions of a <tenant>.user
CREATE TYPE STATUS_TYPE AS ENUM ('valid', 'invalid');

CREATE TABLE sessions
(
    id          VARCHAR(256)                           NOT NULL
        CONSTRAINT sessions_pk
            PRIMARY KEY,                                         -- session id as rand string
    tenant_id   UUID                                   NOT NULL, -- user's tenant's id
    user_id     UUID                                   NOT NULL, -- user id of the user who initiated the session
    occurred_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL, -- first timestamp of the session
    last_event  TIMESTAMP WITH TIME ZONE DEFAULT now(),          -- stores last event's timestamp within this session
    expires_at  TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL, -- session's expiration time
    status      STATUS_TYPE                            NOT NULL  -- session's status
);


