
execute_process(COMMAND go env GOPATH OUTPUT_VARIABLE GOPATH)
string(REGEX REPLACE "\n$" "" GOPATH "${GOPATH}")

include(${GOPATH}/src/github.com/TarsCloud/TarsGo/cmake/tars-tools.cmake)

cmake_minimum_required(VERSION 2.8)

project(CmakeServer Go) # select GO compile

add_subdirectory(client)

gen_server(StressTest CmakeServer)


# go mod init
# mkdir build
# cd build
# cmake ..
# make