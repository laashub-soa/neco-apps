#!/bin/bash

function setup_master() {
    TARGET_POD=$1
    mysqlsh --no-wizard --sql --uri ${MYSQL_OPERATOR_USER}:${MYSQL_OPERATOR_PASSWORD}@${TARGET_POD}.${MYSQL_CLUSTER_DOMAIN}:3306 << EOS
SET GLOBAL rpl_semi_sync_slave_enabled = 0;
SET GLOBAL rpl_semi_sync_master_enabled = 1;
SET GLOBAL offline_mode = 0;
SET GLOBAL rpl_semi_sync_master_timeout = 3600000;
EOS
}

function setup_slave() {
    TARGET_POD=$1
    mysqlsh --no-wizard --sql --uri ${MYSQL_OPERATOR_USER}:${MYSQL_OPERATOR_PASSWORD}@${TARGET_POD}.${MYSQL_CLUSTER_DOMAIN}:3306 << EOS
SET GLOBAL rpl_semi_sync_slave_enabled = 1;
SET GLOBAL rpl_semi_sync_master_enabled = 0;
SET GLOBAL offline_mode = 1;
SET GLOBAL rpl_semi_sync_master_timeout = 3600000;
EOS
}

function change_master_to() {
    TARGET_POD=$1
    master=$2
    mysqlsh --no-wizard --sql --uri ${MYSQL_OPERATOR_USER}:${MYSQL_OPERATOR_PASSWORD}@${TARGET_POD}.${MYSQL_CLUSTER_DOMAIN}:3306 << EOS
CHANGE MASTER TO
  MASTER_HOST = '${master}.${MYSQL_CLUSTER_DOMAIN}',
  MASTER_PORT = 3306,
  MASTER_USER = '${MYSQL_REPLICATION_USER}',
  MASTER_PASSWORD = '${MYSQL_REPLICATION_PASSWORD}',
  MASTER_AUTO_POSITION = 1,
  GET_MASTER_PUBLIC_KEY = 1;
START SLAVE;
EOS
}

function disable_read_only() {
    TARGET_POD=$1
    mysqlsh --no-wizard --sql --uri ${MYSQL_OPERATOR_USER}:${MYSQL_OPERATOR_PASSWORD}@${TARGET_POD}.${MYSQL_CLUSTER_DOMAIN}:3306 << EOS
SET GLOBAL read_only = 0;
EOS
}

setup_master my-app-db-0
setup_slave my-app-db-1
setup_slave my-app-db-2
change_master_to my-app-db-1 my-app-db-0
change_master_to my-app-db-2 my-app-db-0
disable_read_only my-app-db-0
