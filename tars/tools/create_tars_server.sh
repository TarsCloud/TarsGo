#!/bin/sh

# check params
if [ $# -lt 3 ]
then
    echo "<Usage: sh $0  App  Server  Servant>"
        echo ">>>>>>  sh $0  TeleSafe PhonenumSogouServer SogouInfo"
    exit 1
fi

export GOPATH=$(echo $GOPATH | cut -f1 -d ':')
if [ "$GOPATH" == "" ]; then
    echo "GOPATH must be set"
    exit 1
fi

APP=$1
SERVER=$2
SERVANT=$3
TARGET="$GOPATH/src/$APP/$SERVER/"

if [ -d $TARGET ];then
    echo "! Already have some file in $TARGET! Please clear files in prevent of overwrite!"
    exit 1
fi


if [ "$SERVER" == "$SERVANT" ]
then
    echo "Error!(ServerName == ServantName)"
    exit 1
fi
echo "[create server: $APP.$SERVER ...]"

SRC_DIR=$(cd $(dirname $0); pwd)
DEMODIR=$SRC_DIR/Demo
DEBUGDIR=$SRC_DIR/debugtool
cd $DEMODIR || exit 1
SRC_FILE=`find . -maxdepth 1 -type f`

echo "[mkdir: $TARGET]"
mkdir -p $TARGET
cd $TARGET || exit 1

cp -r $DEMODIR/* $TARGET
cp -r $DEBUGDIR $TARGET

if [ `uname` == "Darwin" ] # support macOS
then
    for FILE in $SRC_FILE client/client.go vendor/vendor.json
    do
        echo ">>>Now doing:"$FILE" >>>>"
        sed  -i "" "s/_APP_/$APP/g"   $FILE
        sed  -i "" "s/_SERVER_/$SERVER/g" $FILE
        sed  -i "" "s/_SERVANT_/$SERVANT/g" $FILE
    done

    for RENAMEFILE in `find . -maxdepth 1 -type f`
    do
        # $SERVER cant contain "Servant" string
        NEWFILE=`echo $RENAMEFILE | sed "s/Server/$SERVER/" | sed "s/Servant/$SERVANT/"`
        mv $RENAMEFILE $NEWFILE

        # or use `rename`，default not install rename, you should execute ``` brew install rename ```
        # rename "s/Server/$SERVER/" $RENAMEFILE
        # rename "s/Servant/$SERVANT/" $RENAMEFILE
    done
else
    for FILE in $SRC_FILE client/client.go vendor/vendor.json debugtool/dumpstack.go
    do
        echo ">>>Now doing:"$FILE" >>>>"
        sed  -i "s/_APP_/$APP/g"   $FILE
        sed  -i "s/_SERVER_/$SERVER/g" $FILE
        sed  -i "s/_SERVANT_/$SERVANT/g" $FILE
    done

    for RENAMEFILE in `ls `
    do
        rename "Server" "$SERVER" $RENAMEFILE
        rename "Servant" "$SERVANT" $RENAMEFILE
    done
fi

# try build tars2go
cd "$GOPATH/src/github.com/TarsCloud/TarsGo/tars/tools/tars2go"
go install
cd "$GOPATH/src/$APP/$SERVER"
echo ">>> Great！Done! You can jump in "`pwd`

# show tips: how to convert tars to golang
echo ">>> Tips: After editing the Tars file, execute the following cmd to automatically generate golang files."
echo ">>>       $GOPATH/bin/tars2go *.tars"

