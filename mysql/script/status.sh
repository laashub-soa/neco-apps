#!/bin/bash

SQL_DIR=$(cd $(dirname $0) && pwd)/sql

TARGET_POD=$1

mysqlsh --no-wizard --sql --uri ${MYSQL_OPERATOR_USER}:${MYSQL_OPERATOR_PASSWORD}@${TARGET_POD}.${MYSQL_CLUSTER_DOMAIN}:3306 -f ${SQL_DIR}/status.sql
