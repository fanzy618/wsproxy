#!/bin/sh

wsaddr=${WS_ADDR:=0.0.0.0:1443}

if [[ "$TYPE" = "server" ]]
then
echo "Exec /wsproxy/server -a $a"
exec /wsproxy/server -a $wsaddr

elif [[ "$TYPE" = "client" ]]
then
echo "Exec /wsproxy/client -l $LADDR -r $RADDR -s $SADDR"
exec /wsproxy/client -l $LADDR -r $RADDR -s $SADDR

else
exec "$@"
fi
