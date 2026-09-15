-- gigmatch initial database bootstrap.
-- The MySQL image creates the database from MYSQL_DATABASE; this script makes it
-- idempotent and sets a predictable UTF-8 collation. Tables are created and
-- seeded by the Go backend through GORM AutoMigrate on first boot.
CREATE DATABASE IF NOT EXISTS gigmatch
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

SET NAMES utf8mb4;
SET time_zone = '+08:00';
