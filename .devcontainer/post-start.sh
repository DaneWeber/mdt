#!/usr/bin/env bash
# Runs on every container start (see "postStartCommand" in devcontainer.json).
# Everything here must be idempotent: it re-runs on restart AND after a rebuild.
set -euo pipefail

# 1. Docker seeds a named volume's ownership from the image path it covers, but
#    neither of these paths exists in the image -- so both volumes are created
#    root:root and the "node" user can't write to them. Breaks `claude login`.
sudo chown -R node:node /home/node/.claude /commandhistory

# 2. Install Claude Code from Anthropic's own installer. It puts a launcher at
#    ~/.local/bin/claude pointing into ~/.local/share/claude/versions -- both in
#    the container layer, not on a volume, so a rebuild wipes them and this has
#    to reinstall. A plain restart keeps them, hence the guard: the download is
#    only paid for on a rebuild. Claude Code self-updates in place after that.
#    Must run as "node", not root: the installer installs into $HOME.
if [ ! -x /home/node/.local/bin/claude ]; then
  curl -fsSL https://claude.ai/install.sh | bash
fi

# 3. Point bash history at the /commandhistory volume so it survives rebuilds.
#    ~/.bashrc lives in the image layer and is reset by a rebuild, so re-append
#    the snippet when it's missing. The marker keeps this from stacking up.
MARKER='# >>> devcontainer persistent history >>>'
if ! grep -qF "$MARKER" /home/node/.bashrc; then
  cat >> /home/node/.bashrc <<'SNIPPET'

# >>> devcontainer persistent history >>>
# Keep shell history on the /commandhistory volume so it outlives the container.
export HISTFILE=/commandhistory/.bash_history
# The defaults earlier in this file (1000/2000) truncate a long-lived history.
HISTSIZE=10000
HISTFILESIZE=20000
# Flush after each command, so history survives a hard stop rather than only a
# clean exit. Appended to any existing PROMPT_COMMAND rather than replacing it.
PROMPT_COMMAND="${PROMPT_COMMAND:+$PROMPT_COMMAND; }history -a"
# <<< devcontainer persistent history <<<
SNIPPET
fi

touch /commandhistory/.bash_history
