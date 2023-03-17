#!/bin/bash -p
setup_hashicorp_tool () {
  _tool=${1-NONE}
  case "$1" in
    terraform|packer)
      true;;
    *)
      echo "Not supported" 1>&2
      return 1;;
  esac

  _toolbin=$HOME/.$_tool/bin
  mkdir -p $_toolbin
  pushd $_toolbin 1>/dev/null 2>&1

  # Put $_toolbin at front of $PATH if it's not in $PATH yet
  if ! echo "$PATH" | tr ':' '\n' | grep "$_toolbin" >/dev/null 2>&1 ; then
    NEWPATH="$_toolbin"
    for dir in $(echo $PATH | tr ':' ' ') ; do
      echo "$NEWPATH" | tr ':' '\n' | grep "^${dir}$" > /dev/null 2>&1 || NEWPATH="${NEWPATH}:${dir}"
    done
    export PATH="$NEWPATH"
    unset NEWPATH
  fi

  _toolbase=$(curl https://releases.hashicorp.com/$_tool 2>/dev/null | \
              grep -oP '<a href=".+?">\K.+?(?=<)' | \
              egrep "${_tool}"'_[0-9]+\.[0-9]+\.[0-9]+$' | \
              head -1)

  _toolversion=${_toolbase#${_tool}_}

  _toolbase=${_toolbase}_linux_amd64

  test -f ./$_tool && \
  _installed_toolversion=$(./$_tool --version | egrep -o '[0-9]+\.[0-9]+\.[0-9]+') || \
  _installed_toolversion=None

  echo "Latest $_tool version: $_toolversion"
  echo "Installed $_tool version: $_installed_toolversion"

  if [ "$_toolversion" != "$_installed_toolversion" ] ; then
    echo "updating $_tool:"
    _toolzip=${_toolbase}.zip
    curl -Os https://releases.hashicorp.com/$_tool/$_toolversion/$_toolzip
    unzip $_toolzip
    echo "Installed $_tool, output of ./$_tool --version is:"
    ./$_tool --version
    echo "cleaning up"
    /bin/rm $_toolzip
  else
    echo "$_tool is up to date"
  fi

  # put CWD back
  popd 1>/dev/null 2>&1

  unset _tool
  unset _toolbin
  unset _toolzip
  unset _toolbase
  unset _toolversion
  unset _installed_toolversion
}
