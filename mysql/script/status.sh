#!/bin/bash

SQL_DIR=$(cd $(dirname $0) && pwd)/sql

TARGET_POD=$1

mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@${TARGET_POD}.my-app-db:3306 -f ${SQL_DIR}/status.sql
