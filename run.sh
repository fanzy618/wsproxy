#!/bin/sh

wsaddr=${WS_ADDR:=0.0.0.0:1443}

if [[ "$TYPE" = "server" ]]
then
echo "Exec /wsproxy/wsproxy -role $TYPE -a $wsaddr"
exec /wsproxy/server -role=$TYPE -a $wsaddr

elif [[ "$TYPE" = "client" ]]
then
echo "Exec /wsproxy/wsproxy -role $TYPE  -l $LADDR -r $RADDR -s $SADDR"
exec /wsproxy/wsproxy -role=$TYPE -l $LADDR -r $RADDR -s $SADDR

else
exec "$@"
fi
