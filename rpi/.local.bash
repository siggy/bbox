export GOPATH=$HOME/code/go
export GOROOT=/usr/local/go
export PATH="$PATH:$GOROOT/bin:$GOPATH/bin"

# Increase volume by 5%
alias volu='sudo amixer set Speaker -- $[$(amixer get Speaker|grep -o [0-9]*%|sed 's/%//'|head -n1)+5]%'
# Decrease volume by 5%
alias vold='sudo amixer set Speaker -- $[$(amixer get Speaker|grep -o [0-9]*%|sed 's/%//'|head -n1)-5]%'

alias tj='tmux attach -t sigpi'

tmux new-session -s sigpi -c /home/pi/code/go/src/github.com/siggy/bbox 'bash --init-file <(echo ". \"$HOME/.bashrc\"; sudo ./fish")'
