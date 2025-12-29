# CKS Weight Room .bashrc - matches Killer Coda exam environment
# ~/.bashrc: executed by bash(1) for non-login shells.

# If not running interactively, don't do anything
[ -z "$PS1" ] && return

# History configuration
HISTCONTROL=ignoredups:ignorespace
shopt -s histappend
HISTSIZE=1000
HISTFILESIZE=2000

# Check window size after each command
shopt -s checkwinsize

# Set a simple prompt showing hostname and working directory
export PS1="\[\e]0;\h: \w\a\]\h:\w$ "

# Enable color support for ls and grep
if [ -x /usr/bin/dircolors ]; then
    test -r ~/.dircolors && eval "$(dircolors -b ~/.dircolors)" || eval "$(dircolors -b)"
    alias ls='ls --color=auto'
    alias grep='grep --color=auto'
    alias fgrep='fgrep --color=auto'
    alias egrep='egrep --color=auto'
fi

# Useful ls aliases
alias ll='ls -alF'
alias la='ls -A'
alias l='ls -CF'

# CKS-specific aliases
alias k=kubectl

# Enable bash completion
if [ -f /etc/bash_completion ]; then
    . /etc/bash_completion
fi

# Enable kubectl completion
source <(kubectl completion bash)

# Enable completion for 'k' alias
complete -o default -F __start_kubectl k

# Auto-navigate to home from /root
[[ "$PWD" = /root ]] && cd ~
[[ "$PWD" = /root/Desktop ]] && cd ~
