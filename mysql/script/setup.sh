#!/bin/bash

function setup_master() {
    TARGET_POD=$1
    mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@${TARGET_POD}.my-app-db:3306 << EOS
SET GLOBAL rpl_semi_sync_slave_enabled = 0;
SET GLOBAL rpl_semi_sync_master_enabled = 1;
SET GLOBAL rpl_semi_sync_master_timeout = 3600000;
EOS
}

function setup_slave() {
    TARGET_POD=$1
    mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@${TARGET_POD}.my-app-db:3306 << EOS
SET GLOBAL rpl_semi_sync_slave_enabled = 1;
SET GLOBAL rpl_semi_sync_master_enabled = 0;
SET GLOBAL rpl_semi_sync_master_timeout = 3600000;
EOS
}

function change_master_to() {
    TARGET_POD=$1
    master=$2
    mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@${TARGET_POD}.my-app-db:3306 << EOS
CHANGE MASTER TO MASTER_HOST = '${master}.my-app-db', MASTER_PORT = 3306, MASTER_USER = 'root', MASTER_PASSWORD = 'cybozu', MASTER_AUTO_POSITION = 1;
START SLAVE;
EOS
}

function disable_super_read_only() {
    TARGET_POD=$1
    mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@${TARGET_POD}.my-app-db:3306 << EOS
SET @@GLOBAL.SUPER_READ_ONLY=OFF;
EOS
}

setup_master my-app-db-0
setup_slave my-app-db-1
setup_slave my-app-db-2
change_master_to my-app-db-1 my-app-db-0
change_master_to my-app-db-2 my-app-db-0
disable_super_read_only my-app-db-0
