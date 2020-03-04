#!/bin/bash

mysqlsh --no-wizard --py --uri root:${MYSQL_ROOT_PASSWORD}@my-app-db-0.my-app-db:3306 -e "dba.create_cluster('Cluster', {'ipWhitelist': '10.0.0.0/8'})"
mysqlsh --no-wizard --py --uri root:${MYSQL_ROOT_PASSWORD}@my-app-db-0.my-app-db:3306 -e "dba.get_cluster('Cluster').add_instance('root:${MYSQL_ROOT_PASSWORD}@my-app-db-1.my-app-db:3306', {'ipWhitelist': '10.0.0.0/8', 'recoveryMethod': 'clone'})"
mysqlsh --no-wizard --py --uri root:${MYSQL_ROOT_PASSWORD}@my-app-db-0.my-app-db:3306 -e "dba.get_cluster('Cluster').add_instance('root:${MYSQL_ROOT_PASSWORD}@my-app-db-2.my-app-db:3306', {'ipWhitelist': '10.0.0.0/8', 'recoveryMethod': 'clone'})"
