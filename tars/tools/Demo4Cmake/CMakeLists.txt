
execute_process(COMMAND go env GOPATH OUTPUT_VARIABLE GOPATH)
string(REGEX REPLACE "\n$" "" GOPATH "${GOPATH}")

include(${GOPATH}/src/github.com/TarsCloud/TarsGo/cmake/tars-tools.cmake)

cmake_minimum_required(VERSION 2.8)

project(_SERVER_ Go) # select GO compile

gen_server(_APP_ _SERVER_)

add_subdirectory(client)

# go env -w GO111MODULE=on
# go mod init
# mkdir build
# cd build
# cmake ..
# make