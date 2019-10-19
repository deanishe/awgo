#!/usr/bin/env zsh

root="$( git rev-parse --show-toplevel )"
testdir="${root}/testenv"
iplist="${root}/info.plist"
covfile="${root}/coverage.out"
covjson="${root}/coverage.json"
covhtml="${root}/coverage.html"

verbose=false
runlint=false
runtests=true
opencover=false
usegocov=false
cover=false
mkip=false
vopt=
gopts=()

# log <arg>... | Echo arguments to STDERR
log() {
  echo "$@" >&2
}

# installed <prog> | Check whether program is installed
installed() {
  hash "$1" &>/dev/null
  return $?
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
    run-tests.sh [-v|-V] [-c] [-C] [-i] [-g]
    run-tests.sh -l
    run-tests.sh [-g] -r
    run-tests.sh -h

Options:
    -c      Write coverage report
    -C      Open HTML coverage report
    -l      Lint project
    -r      Just open coverage report
    -g      Use gocov for coverage report (implies -c)
    -i      Create a dummy info.plist
    -h      Show this help message and exit
    -v      Be verbose
    -V      Be even more verbose
EOF
}

while getopts ":CcghilrvV" opt; do
  case $opt in
    c)
      cover=true
      ;;
    g)
      usegocov=true
      cover=true
      ;;
    C)
      opencover=true
      cover=true
      ;;
    i)
      mkip=true
      ;;
    l)
      runlint=true
      runtests=false
      ;;
    r)
      opencover=true
      runtests=false
      ;;
    V)
      gopts+=(-v)
      verbose=true
      vopt='-v'
      ;;
    v)
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

$runlint && {
  golangci-lint run -c .golangci.toml
  st=$?
  [[ $st -ne 0 ]] && {
    fail "linting failed"
  }
  success "linting passed"
  exit 0
}

$cover && gopts+=(-coverprofile="$covfile")

command mkdir $vopt -p "${testdir}"/{data,cache}
$mkip && touch $vopt "$iplist"

cd "$root"
source "env.sh"
export alfred_version=
export alfred_workflow_data="${testdir}/data"
export alfred_workflow_cache="${testdir}/cache"

[[ $#@ -eq 0 ]] && {
  pkgs=(./...)
} || {
  pkgs=($@)
}

st=0
$runtests && {
#  go test $gopts $pkgs
  go test -cover -json $gopts $pkgs | go run github.com/mfridman/tparse
#  gotestsum -- $gopts $pkgs
  st=$?

  [[ $st -eq 0 ]] && {
    success "passed"
  }
  command rm $vopt -rf "$testdir"/*
}

test -f "$iplist" && command rm $vopt -f "$iplist"

cd -

[[ $st -ne 0 ]] && {
  fail "failed"
}

$opencover && {
  $usegocov && {
    gocov convert "$covfile" > "$covjson"
    gocov-html > "$covhtml" < "$covjson"
    open "$covhtml"
  } || {
    go tool cover -html="$covfile"
  }
}

# $cover && installed gocov && {
#   gocov convert "$covfile" > "$covjson"
#   installed gocov-html && {
#     gocov-html > "$covhtml" < "$covjson"
#   }
# }

exit 0

#  vim: set ft=zsh ts=2 sw=2 tw=100 et :
