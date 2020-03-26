#!/bin/bash

mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@my-app-db-0.my-app-db:3306 -f ./sql/setup_master.sql
mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@my-app-db-1.my-app-db:3306 -f ./sql/setup_slave.sql
mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@my-app-db-2.my-app-db:3306 -f ./sql/setup_slave.sql
