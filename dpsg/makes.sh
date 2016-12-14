#!/bin/sh
#
setverson() {
    svnv=`svnversion |sed 's/^.*://' |sed 's/[A-Z]*$//'`
    svnv=21
    echo $svnv

    regs="s/VersionNew\(.*\)[0-9]/VersionNew\1$svnv/"
    sed -i $regs client/dpsg/Resources/x-Version.xml
    sed -i $regs bin/x-Version.json
    #cat x-Version.xml |grep VersionNew

    makes1
}
resetverson() {
    svnv=`svnversion |sed 's/^.*://' |sed 's/[A-Z]*$//'`
    svnv=32323
    echo $svnv

    regs="s/VersionOld\(.*\)[0-9]/VersionOld\1$svnv/"
    sed -i $regs client/dpsg/Resources/x-Version.xml
    sed -i $regs bin/x-Version.json
    #cat x-Version.xml |grep versionOld

    makes1
}

makes1(){

        make
	export GOROOT=/usr/local/go
	export GOBIN=$GOROOT/bin
	export GOPATH=$HOME/dev/dpsg/dpsg/server/3rdpkg:$HOME/dev/dpsg/dpsg/server:$GOROOT
	export PATH=$PATH:$GOROOT/bin:$GOBIN
	cd bin
	go build tools/cnserver
	echo build cnserver ok !
	go build tools/dbserver
	echo build dbserver ok !
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
	export GOPATH=$HOME/dev/dpsg/dpsg/server/tools/GmTools:$GOPATH
	go build -o ../server/tools/GmTools/gmtools/gmtools ../server/tools/GmTools/gmtools/main.go
	echo build gmtools ok!

}


case "$1" in
    v)
        setverson
        ;;
    V)
        resetverson
        ;;
    *)
        echo $"Usage: $0 {v (unSetVerson)|V (reSetVerson)}"
        exit 2
esac
