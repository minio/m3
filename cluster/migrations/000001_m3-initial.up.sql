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

create table admins
(
    id               uuid
        constraint admins_pk
            primary key,
    name             varchar(256),
    email            varchar(256)              not null,
    password         varchar(256),
    sys_created_by   varchar(256)              not null,
    sys_created_date timestamptz default now() not null
);

create unique index admins_email_uindex
    on admins (email);

create table tenants
(
    id               uuid                                   not null
        constraint tenants_pk
            primary key,
    name             varchar(256)                           not null,
    short_name       varchar(256)                           not null,
    sys_created_by   varchar(256)                           not null,
    sys_created_date timestamp with time zone default now() not null
);


create table storage_groups
(
    id               uuid                                   not null
        constraint storage_groups_pk
            primary key,
    name             varchar(256),
    num              serial                                 not null,
    sys_created_by   varchar(256)                           not null,
    sys_created_date timestamp with time zone default now() not null
);


create table tenants_storage_groups
(
    tenant_id        uuid                                   not null
        constraint tenants_storage_groups_tenants_id_fk
            references tenants,
    storage_group_id uuid                                   not null
        constraint tenants_storage_groups_storage_groups_id_fk
            references storage_groups,
    port             integer                                not null,
    service_name     varchar(64)                            not null,
    sys_created_by   varchar(256)                           not null,
    sys_created_date timestamp with time zone default now() not null
);


create table nodes
(
    id               uuid                                   not null
        constraint nodes_pk
            primary key,
    name             varchar(256),
    k8s_label        varchar(256),
    sys_created_by   varchar(256)                           not null,
    sys_created_date timestamp with time zone default now() not null
);


create table storage_clusters
(
    id               uuid                                   not null
        constraint storage_clusters_pk
            primary key,
    name             varchar(256),
    sys_created_by   varchar(256)                           not null,
    sys_created_date timestamp with time zone default now() not null
);


create table storage_clusters_groups
(
    storage_cluster_id uuid                                   not null
        constraint storage_clusters_groups_storage_clusters_id_fk
            references storage_clusters,
    storage_group_id   uuid                                   not null
        constraint storage_clusters_groups_storage_groups_id_fk
            references storage_groups,
    sys_created_by     varchar(256)                           not null,
    sys_created_date   timestamp with time zone default now() not null
);


create table storage_cluster_nodes
(
    storage_cluster_id uuid                                   not null
        constraint storage_cluster_nodes_storage_clusters_id_fk
            references storage_clusters,
    node_id            uuid                                   not null
        constraint storage_cluster_nodes_nodes_id_fk
            references nodes,
    k8s_label          varchar(256),
    sys_created_by     varchar(256)                           not null,
    sys_created_date   timestamp with time zone default now() not null
);


create table node_volumes
(
    id               uuid                                   not null
        constraint node_volumes_pk
            primary key,
    node_id          uuid                                   not null
        constraint node_volumes_nodes_id_fk
            references nodes,
    mount_path       varchar(256),
    sys_created_by   varchar(256)                           not null,
    sys_created_date timestamp with time zone default now() not null
);

--  Table to store Disks attached to a node and their mount points

create table disks
(
    id               uuid                                   not null
        constraint disks_pk
            primary key,
    node_id          uuid
        constraint disks_nodes_id_fk
            references nodes,
    mount_point      varchar(512),
    capacity         bigint,
    sys_created_by   varchar(256)                           not null,
    sys_created_date timestamp with time zone default now() not null
);

comment on column disks.capacity is 'Capacity in bytes';

--  Table to store sessions of a <tenant>.user
CREATE TYPE status_type AS ENUM ('valid', 'invalid');

create table sessions
(
    id          varchar(256)                           not null
        constraint sessions_pk
            primary key,                                         -- session id as rand string
    tenant_id   uuid                                   not null, -- user's tenant's id
    user_id     uuid                                   not null, -- user id of the user who initiated the session
    occurred_at timestamp with time zone default now() not null, -- first timestamp of the session
    last_event  timestamp with time zone default now(),          -- stores last event's timestamp within this session
    expires_at  timestamp with time zone default now() not null, -- session's expiration time
    status      status_type                            not null  -- session's status
);


