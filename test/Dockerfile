FROM golang:1.25-bookworm

SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

# ── System packages ──────────────────────────────────────────────────────────
RUN apt-get update && apt-get install -y --no-install-recommends \
        curl \
        ca-certificates \
        git \
        shellcheck \
        python3 \
        python3-pip \
        unzip \
    && rm -rf /var/lib/apt/lists/*

# ── Node.js 20 (stylelint 17 requires Node >= 20) ─────────────────────────────
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

# ── uv (Python tool manager) ─────────────────────────────────────────────────
RUN curl -LsSf https://astral.sh/uv/install.sh | sh
ENV PATH="/root/.local/bin:$PATH"

# ── Python tools ─────────────────────────────────────────────────────────────
RUN uv tool install ruff && \
    uv tool install mypy && \
    uv tool install bandit && \
    uv tool install pylint && \
    uv tool install basedpyright && \
    uv tool install semgrep

# ── Node.js tools ─────────────────────────────────────────────────────────────
RUN npm install -g \
        eslint@9 \
        @biomejs/biome \
        stylelint \
        stylelint-config-standard

# ── hadolint ─────────────────────────────────────────────────────────────────
RUN ARCH=$(dpkg --print-architecture) && \
    case "$ARCH" in \
        amd64) HADOLINT_ARCH="x86_64" ;; \
        arm64) HADOLINT_ARCH="arm64" ;; \
        *) echo "Unsupported arch: $ARCH" && exit 1 ;; \
    esac && \
    curl -fsSL "https://github.com/hadolint/hadolint/releases/download/v2.12.0/hadolint-Linux-${HADOLINT_ARCH}" \
        -o /usr/local/bin/hadolint && \
    chmod +x /usr/local/bin/hadolint

# ── golangci-lint ─────────────────────────────────────────────────────────────
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
        | sh -s -- -b /usr/local/bin v1.61.0

# ── trivy ─────────────────────────────────────────────────────────────────────
RUN curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh \
        | sh -s -- -b /usr/local/bin

# ── Rust + cargo-clippy ───────────────────────────────────────────────────────
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \
        | sh -s -- -y --profile minimal --component clippy
ENV PATH="/root/.cargo/bin:$PATH"

# ── Project source ────────────────────────────────────────────────────────────
WORKDIR /project
COPY . .

# Build all prettyout plugins
RUN go build -o /usr/local/bin/prettyout-ruff          ./cmd/prettyout-ruff/ && \
    go build -o /usr/local/bin/prettyout-basedpyright  ./cmd/prettyout-basedpyright/ && \
    go build -o /usr/local/bin/prettyout-mypy           ./cmd/prettyout-mypy/ && \
    go build -o /usr/local/bin/prettyout-bandit         ./cmd/prettyout-bandit/ && \
    go build -o /usr/local/bin/prettyout-pylint         ./cmd/prettyout-pylint/ && \
    go build -o /usr/local/bin/prettyout-eslint         ./cmd/prettyout-eslint/ && \
    go build -o /usr/local/bin/prettyout-biome          ./cmd/prettyout-biome/ && \
    go build -o /usr/local/bin/prettyout-stylelint      ./cmd/prettyout-stylelint/ && \
    go build -o /usr/local/bin/prettyout-shellcheck     ./cmd/prettyout-shellcheck/ && \
    go build -o /usr/local/bin/prettyout-hadolint       ./cmd/prettyout-hadolint/ && \
    go build -o /usr/local/bin/prettyout-golangci       ./cmd/prettyout-golangci/ && \
    go build -o /usr/local/bin/prettyout-cargo-clippy   ./cmd/prettyout-cargo-clippy/ && \
    go build -o /usr/local/bin/prettyout-trivy          ./cmd/prettyout-trivy/ && \
    go build -o /usr/local/bin/prettyout-semgrep        ./cmd/prettyout-semgrep/

WORKDIR /test-workspace
CMD ["/project/test/run.sh"]
