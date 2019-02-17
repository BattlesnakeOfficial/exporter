#!/bin/bash
rm -rf /tmp/board

mkdir /tmp/board
wget -qO /tmp/board/master.zip "https://github.com/battlesnakeio/board/archive/master.zip" 
unzip -q /tmp/board/master -d /tmp/board

rm -rf snake-images
mkdir snake-images
cp -rv /tmp/board/board-master/public/images/snake/* snake-images