#!/bin/bash
echo "正在部署服务器管理台..."
# 部署服务
/bin/cp -rf ./siteweb-manager /etc/init.d/siteweb-manager
chmod +x /etc/init.d/siteweb-manager
chmod +x /siteweb/siteweb-manager/siteweb-manager

echo "正在初始化服务器管理台..."
# 初始化数据库
if [ ! -f /siteweb/siteweb-manager/sqlite.db ]; then
    DIR=$(pwd)
    cd /siteweb/siteweb-manager
    siteweb-manager migrate -c /siteweb/siteweb-manager/config/settings.yml
    cd $DIR
fi

echo "正在启动服务器管理台..."
chkconfig --add siteweb-manager
service siteweb-manager restart
# chkconfig --level 2345 siteweb-manager on

# 软链 全局命令
ln -s -f /siteweb/siteweb-manager/siteweb-manager  /usr/bin/siteweb-manager
