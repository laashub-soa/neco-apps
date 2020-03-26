INSTALL PLUGIN rpl_semi_sync_slave SONAME 'semisync_slave.so';
SET GLOBAL rpl_semi_sync_slave_enabled = 1;
CHANGE MASTER TO MASTER_HOST = 'my-app-db-0.my-app-db', MASTER_PORT = 3306, MASTER_USER = 'root', MASTER_PASSWORD = 'cybozu', MASTER_AUTO_POSITION = 1;
START SLAVE;
