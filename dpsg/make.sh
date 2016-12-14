make
export GOROOT=/usr/local/go
export GOBIN=$GOROOT/bin
export GOPATH=$HOME/dev/dpsg/dpsg/server/3rdpkg:$HOME/dev/dpsg/dpsg/server:$GOROOT
export PATH=$PATH:$GOROOT/bin:$GOBIN
cd bin
go build tools/gameserver
echo build gameserver ok !
go build tools/dbserver
echo build dbserver ok !
go build tools/accountserver
echo build accountserver ok !
go build tools/gateserver
echo build gateserver ok !
go build tools/center
echo build center ok !
go build tools/logserver
echo build logserver ok !
go build tools/chatserver
echo build chatserver ok !
go build tools/gmserver
echo build gmserver ok!
go build tools/lockserver
echo build lockserver ok!
export GOPATH=$HOME/dev/dpsg/dpsg/server/tools/GmTools:$GOPATH
go build -o ../server/tools/GmTools/gmtools/gmtools ../server/tools/GmTools/gmtools/main.go
echo build gmtools ok!