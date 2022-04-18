# cat /etc/bash_completion.d/keentune.bash   
_keentune()  
{
    COMPREPLY=()
    local cur=${COMP_WORDS[COMP_CWORD]};
    local cmd=${COMP_WORDS[COMP_CWORD-1]};
    case $cmd in
    'keentune')
          COMPREPLY=( $(compgen -W '--help param profile sensitize version' -- $cur) ) ;;
    'param')
          COMPREPLY=( $(compgen -W 'delete dump jobs list rollback stop tune' -- $cur) ) ;;
    'tune')
          COMPREPLY=( $(compgen -W '--job' -- $cur) ) ;;
    'profile')
          COMPREPLY=( $(compgen -W 'delete generate info list rollback set' -- $cur) ) ;;
    'senisitize')
          COMPREPLY=( $(compgen -W 'collect delete list stop train' -- $cur) ) ;;  
    'set')
          COMPREPLY=( $(compgen -W '--group1' -- $cur) ) ;; 
    esac


    if [[ "${COMP_WORDS[3]}" == "--group1" && "${COMP_WORDS[2]}" == "set" ]]; then
	    local pro=($(pwd))
	    cd /etc/keentune/profile
	    compopt -o nospace
	    COMPREPLY=($(compgen -d -f -- $cur))
	    cd $pro
    fi

    return 0
} 
complete -F _keentune keentune

