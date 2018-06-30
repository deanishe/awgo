#!/usr/bin/env zsh

# URL of icon generator
root="$( cd "$( dirname "$0" )"; pwd )"
testdir="${root}/testenv"
iplist="${root}/info.plist"
covfile="${root}/coverage.out"

verbose=false
opencover=false
cover=false
mkip=false
vopt=
gopts=()

# log <arg>... | Echo arguments to STDERR
log() {
  echo "$@" >&2
}

# info <arg>.. | Write args to STDERR if VERBOSE is true
info() {
  $verbose && log $(print -P "%F{blue}.. %f") "$@"
  return 0
}

# success <arg>.. | Write green "ok" and args to STDERR if VERBOSE is true
success() {
  # $verbose && log $(print -P "%F{green}ok %f") "$@"
  log $(print -P "%F{green}#####################################%f")
  log $(print -P "%F{green}# $@ %f")
  log $(print -P "%F{green}#####################################%f")
  return 0
}

# error <arg>.. | Write red "error" and args to STDERR
error() {
  log $(print -P '%F{red}err%f') "$@"
}

# fail <arg>.. | Write red "error" and args to STDERR, then exit with status 1
fail() {
  log $(print -P "%F{red}#####################################%f")
  log $(print -P "%F{red}# $@ %f")
  log $(print -P "%F{red}#####################################%f")
  # error "$@"
  exit 1
}

usage() {
cat <<EOF
run-tests.sh [options] [<module>...]

Run unit tests in a workflow-like environment.

Usage:
    run-tests.sh [-v|-V] [-c] [-i] [-H]
    run-tests.sh -h

Options:
    -c      Write coverage report
    -i      Create a dummy info.plist
    -H      Open HTML coverage report
    -h      Show this help message and exit
    -v      Be verbose
    -V      Be even more verbose
EOF
}

while getopts ":HchivV" opt; do
  case $opt in
    c)
      cover=true
      ;;
    H)
      opencover=true
      cover=true
      ;;
    i)
      mkip=true
      ;;
    V)
      gopts+=(-v)
      verbose=true
      vopt='-v'
      ;;
    V)
      gopts+=(-v)
      verbose=true
      ;;
    h)
      usage
      exit 0
      ;;
    \?)
      fail "invalid option: -$OPTARG";;
  esac
done
shift $((OPTIND-1))

$cover && gopts+=(-coverprofile="$covfile")

command mkdir $vopt -p "${testdir}"/{data,cache}
$mkip && touch $vopt "$iplist"

# Absolute bare-minimum for AwGo to function...
export alfred_workflow_bundleid="net.deanishe.awgo"
export alfred_workflow_cache="${testdir}/cache"
export alfred_workflow_data="${testdir}/data"

# Expected by ExampleNew
export alfred_workflow_version="0.14"
export alfred_workflow_name="AwGo"

cd "$root"

go test $gopts "$@"
st=$?

cd -

command rm $vopt -rf "$testdir"/*
test -f "$iplist" && command rm $vopt -f "$iplist"

[[ st -eq 0 ]] || {
  fail "go test failed with $st"
}

success "tests passed"

$opencover && go tool cover -html="$covfile"

exit 0

#  vim: set ft=zsh ts=2 sw=2 tw=100 et :
