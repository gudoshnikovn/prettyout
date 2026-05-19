package completion

import (
	"fmt"

	"github.com/gudoshnikov_na/prettyout/internal/registry"
)

// Generate returns a shell completion script for prettyout.
// Tool names are fetched dynamically via `prettyout _completions tools` at completion time.
func Generate(shell string, _ *registry.Registry) (string, error) {
	switch shell {
	case "zsh":
		return generateZsh(), nil
	case "bash":
		return generateBash(), nil
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: zsh, bash)", shell)
	}
}

func generateZsh() string {
	return `_prettyout() {
  local -a cmds
  cmds=(setup hook run enable disable list install update upgrade doctor status completion)
  if (( CURRENT == 2 )); then
    _describe 'command' cmds
    return
  fi
  case "$words[2]" in
    run)
      if (( CURRENT == 3 )); then
        local -a runflags
        runflags=(--raw '--group-by=rule' '--group-by=file' '--sort=count' '--sort=alpha' '--only-rules=' '--only-files=')
        _describe 'flag' runflags
      fi
      ;;
    enable|disable|install|update)
      local -a tools
      tools=(${(f)"$(prettyout _completions tools 2>/dev/null)"})
      _describe 'tool' tools
      ;;
  esac
}
compdef _prettyout prettyout
`
}

func generateBash() string {
	return `_prettyout() {
  local cur="${COMP_WORDS[COMP_CWORD]}"
  local prev="${COMP_WORDS[COMP_CWORD-1]}"
  if [[ $COMP_CWORD -eq 1 ]]; then
    COMPREPLY=($(compgen -W "setup hook run enable disable list install update upgrade doctor status completion" -- "$cur"))
    return
  fi
  case "$prev" in
    enable|disable|install|update)
      local tools
      tools=$(prettyout _completions tools 2>/dev/null)
      COMPREPLY=($(compgen -W "$tools" -- "$cur"))
      ;;
  esac
}
complete -F _prettyout prettyout
`
}
