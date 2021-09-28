cmake_minimum_required(VERSION 2.8...3.20)
project(TDengine)

set(TARGET_OUT blm_demo)

# Go configurations

set(GO_SRCS
  main.go)

# Custom command for 'go build -buildmode=c-archive ...'
# to create a library from Go codes.
add_custom_command(OUTPUT ${TARGET_OUT}
  DEPENDS ${GO_SRCS}
  WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
  COMMAND env go build
  COMMENT "Building BLM3")

# A library and a header are generated in the binary directory.
include_directories(${CMAKE_CURRENT_BINARY_DIR})
link_directories(${CMAKE_CURRENT_BINARY_DIR})

add_executable(${TARGET_OUT})