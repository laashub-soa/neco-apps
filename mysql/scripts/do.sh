#!/bin/bash

ACTION=$1
TARGET=$2

function create_table() {
    node=$1
    mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@${node}.my-app-db:3306 << EOS
CREATE DATABASE IF NOT EXISTS test;
CREATE TABLE IF NOT EXISTS test.t1 (
  num  bigint unsigned NOT NULL AUTO_INCREMENT,
  val0 varchar(100) DEFAULT NULL,
  val1 varchar(100) DEFAULT NULL,
  val2 varchar(100) DEFAULT NULL,
  val3 varchar(100) DEFAULT NULL,
  val4 varchar(100) DEFAULT NULL,
  UNIQUE KEY num (num)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
EOS
}

function insert() {
    node=$1
    mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@${node}.my-app-db:3306 << EOS
SET autocommit=0;
START TRANSACTION;
INSERT INTO test.t1 (val0, val1, val2, val3, val4) values
('ZYeMhKcUOQjLGSxXHHDcjqRP2fVnou6S', 'tMMAoflfoDC3UXaiPiPp3vVljwcCRuAz', '9VMkikrF5bPlxyiLGo9KOFJWNxewTgWd', 'SB3lrOXaIkqdyCgfZ0q7pbEWg2ZIVzGg', 'T5NlRPKyn3FH2dQJcNbUcax0V5efjhbv'),
('ZYeMhKcUOQjLGSxXHHDcjqRP2fVnou6S', 'tMMAoflfoDC3UXaiPiPp3vVljwcCRuAz', '9VMkikrF5bPlxyiLGo9KOFJWNxewTgWd', 'SB3lrOXaIkqdyCgfZ0q7pbEWg2ZIVzGg', 'T5NlRPKyn3FH2dQJcNbUcax0V5efjhbv'),
('ZYeMhKcUOQjLGSxXHHDcjqRP2fVnou6S', 'tMMAoflfoDC3UXaiPiPp3vVljwcCRuAz', '9VMkikrF5bPlxyiLGo9KOFJWNxewTgWd', 'SB3lrOXaIkqdyCgfZ0q7pbEWg2ZIVzGg', 'T5NlRPKyn3FH2dQJcNbUcax0V5efjhbv'),
('ZYeMhKcUOQjLGSxXHHDcjqRP2fVnou6S', 'tMMAoflfoDC3UXaiPiPp3vVljwcCRuAz', '9VMkikrF5bPlxyiLGo9KOFJWNxewTgWd', 'SB3lrOXaIkqdyCgfZ0q7pbEWg2ZIVzGg', 'T5NlRPKyn3FH2dQJcNbUcax0V5efjhbv'),
('ZYeMhKcUOQjLGSxXHHDcjqRP2fVnou6S', 'tMMAoflfoDC3UXaiPiPp3vVljwcCRuAz', '9VMkikrF5bPlxyiLGo9KOFJWNxewTgWd', 'SB3lrOXaIkqdyCgfZ0q7pbEWg2ZIVzGg', 'T5NlRPKyn3FH2dQJcNbUcax0V5efjhbv');
COMMIT;
EOS
}

function count() {
    node=$1
    mysqlsh --no-wizard --sql --uri root:${MYSQL_ROOT_PASSWORD}@${node}.my-app-db:3306 << EOS
SELECT count(*) FROM test.t1;
EOS
}

case "${ACTION}" in
    "create")
        create_table ${TARGET} ;;
    "insert")
        insert ${TARGET} ;;
    "count")
        count ${TARGET} ;;
    *)
        echo "invalid action: $0 <create|insert|count> <mysql host>";;
esac
