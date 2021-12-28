package db

import (
    "github.com/go-sql-driver/mysql"
)

//mariaDB db_user Config
var Cfg = mysql.Config{
        User:   "WilsonLo",
        Passwd:  "Xjij0vu;;",
        Net:    "tcp",
        Addr:   "127.0.0.1:3306",
        DBName: "db_user",
        AllowNativePasswords: true,
}

//mariaDB db_device Config
var Cfg_device = mysql.Config{
        User:   "WilsonLo",
        Passwd:  "Xjij0vu;;",
        Net:    "tcp",
        Addr:   "127.0.0.1:3306",
        DBName: "db_device",
        AllowNativePasswords: true,
}