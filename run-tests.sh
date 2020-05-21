#!/usr/bin/env zsh

root="$( git rev-parse --show-toplevel )"
testdir="${root}/testenv"
iplist="${root}/info.plist"
covfile="${root}/coverage.out"
covhtml="${root}/coverage.html"

verbose=false
runlint=false
runtests=true
opencover=false
usegocov=false
cover=false
mkip=false
colour=false
vopt=
gopts=()

test -t 1 && colour=true

# log <arg>... | Echo arguments to STDERR
log() {
  echo "$@" >&2
}

# installed <prog> | Check whether program is installed
installed() {
  hash "$1" &>/dev/null
  return $?
}

# success <arg>... | Write message in green to STDOUT
success() {
  $verbose || return 0
  $colour && {
    print -P "%F{green}$@ %f"
  } || echo "[OK]  $@"
}

# error <arg>... | Write message in red to STDERR
error() {
  $colour && {
    print -P "%F{red}$@ %f" >&2
  } || echo "[ERR] $@" >&2
}

# fail <arg>... | Write message in red to STDERR, then exit with status 1
fail() {
  error "$@"
  exit 1
}

usage() {
cat <<EOF
run-tests.sh [options] [<module>...]

Run unit tests in a workflow-like environment.

Usage:
    run-tests.sh [-v|-V] [-t] [-c|-g] [-C] [-i]
    run-tests.sh [-t] -l
    run-tests.sh [-g] -r
    run-tests.sh -h

Options:
    -c      write coverage report
    -C      open HTML coverage report
    -l      lint project
    -r      just open coverage report
    -g      use gocov for coverage report (implies -c)
    -i      create a dummy info.plist
    -t      force terminal (coloured) output
    -h      show this help message and exit
    -v      be verbose
    -V      be even more verbose
EOF
}

while getopts ":CcghilrtvV" opt; do
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
    t)
      colour=true
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
  diff=($(gofmt -s -l **/*.go))
  test -z "$diff" || {
    for s in $diff; do error "bad formatting: $s"; done
    fail "gofmt -s found incorrectly formatted files"
  }
  success "all files formatted correctly"

  go run golang.org/x/lint/golint -set_exit_status ./...
  [[ $? -eq 0 ]] || fail "linting with golint failed"
  success "golint found no issues"

  go run github.com/golangci/golangci-lint/cmd/golangci-lint run -c .golangci.toml
  [[ $? -eq 0 ]] || fail "linting with golangci-lint failed"
  success "golangci-lint found no issues"
  exit 0
}

$cover && gopts+=(-coverprofile="$covfile")

command mkdir $vopt -p "${testdir}"/{data,cache}
$mkip touch $vopt "$iplist"
trap "test -f \"$iplist\" && rm -f \"$iplist\"; test -d \"$testdir\" && rm -rf \"$testdir\";" EXIT INT TERM

cd "$root"
source "env.sh"
export alfred_version=
export alfred_workflow_data="${testdir}/data"
export alfred_workflow_cache="${testdir}/cache"

pkgs=(./...)
[[ $#@ -eq 0 ]] || pkgs=($@)

st=0
$runtests && {
  go test -cover -json $gopts $pkgs | go run github.com/mfridman/tparse
#  gotestsum -- $gopts $pkgs
  st=$?
  [[ $st -eq 0 ]] && success "unit tests passed"
}

cd -

[[ $st -ne 0 ]] && fail "unit tests failed"

$opencover && {
  $usegocov && {
    go run github.com/axw/gocov/gocov convert "$covfile" \
    | go run github.com/matm/gocov-html > "$covhtml"
    open "$covhtml"
  } || {
    go tool cover -html="$covfile"
  }
}

exit 0

#  vim: set ft=zsh ts=2 sw=2 tw=100 et :
