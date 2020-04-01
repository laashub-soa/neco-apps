#!/bin/bash

SQL_DIR=$(cd $(dirname $0) && pwd)/sql

function setup_node() {
    node=$1
    mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@${node}.my-app-db:3306 -f ${SQL_DIR}/setup_node.sql
}

function change_master_to() {
    node=$1
    master=$2
    mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@${node}.my-app-db:3306 << EOS
CHANGE MASTER TO MASTER_HOST = '${master}.my-app-db', MASTER_PORT = 3306, MASTER_USER = 'root', MASTER_PASSWORD = 'cybozu', MASTER_AUTO_POSITION = 1;
START SLAVE;
EOS
}

function disable_super_read_only() {
    node=$1
    mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@${node}.my-app-db:3306 << EOS
SET @@GLOBAL.SUPER_READ_ONLY=OFF;
EOS
}

setup_node my-app-db-0
setup_node my-app-db-1
setup_node my-app-db-2
change_master_to my-app-db-1 my-app-db-0
change_master_to my-app-db-2 my-app-db-0
disable_super_read_only my-app-db-0
