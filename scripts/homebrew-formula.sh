#!/bin/sh
# Render a Homebrew formula for dataflow-cli to stdout.
# Usage: homebrew-formula.sh VERSION DARWIN_ARM64_SHA DARWIN_AMD64_SHA LINUX_ARM64_SHA LINUX_AMD64_SHA
# VERSION is X.Y.Z without a leading v. Each SHA is the sha256 of the matching .tar.gz asset.
set -eu

if [ "$#" -ne 5 ]; then
  echo "usage: homebrew-formula.sh VERSION DARWIN_ARM64_SHA DARWIN_AMD64_SHA LINUX_ARM64_SHA LINUX_AMD64_SHA" >&2
  exit 1
fi

version=$1
darwin_arm64=$2
darwin_amd64=$3
linux_arm64=$4
linux_amd64=$5

case "$version" in
  ''|*[!0-9.]*|*.*.*.*)
    echo "version must be X.Y.Z without a leading v" >&2
    exit 1
    ;;
esac
case "$version" in
  *.*.*) ;;
  *)
    echo "version must be X.Y.Z without a leading v" >&2
    exit 1
    ;;
esac

check_sha() {
  if ! printf '%s' "$1" | grep -Eq '^[0-9a-f]{64}$'; then
    echo "sha256 must be 64 lowercase hex characters" >&2
    exit 1
  fi
}

check_sha "$darwin_arm64"
check_sha "$darwin_amd64"
check_sha "$linux_arm64"
check_sha "$linux_amd64"

base="https://github.com/dataflow-operator/dataflow-cli/releases/download/v${version}"

cat <<EOF
class DataflowCli < Formula
  desc "CLI for the DataFlow operator"
  homepage "https://dataflow-operator.github.io/docs/cli/"
  version "${version}"
  license "Apache-2.0"

  on_macos do
    on_arm do
      url "${base}/dataflow-cli-darwin-arm64.tar.gz"
      sha256 "${darwin_arm64}"
    end
    on_intel do
      url "${base}/dataflow-cli-darwin-amd64.tar.gz"
      sha256 "${darwin_amd64}"
    end
  end

  on_linux do
    on_arm do
      url "${base}/dataflow-cli-linux-arm64.tar.gz"
      sha256 "${linux_arm64}"
    end
    on_intel do
      url "${base}/dataflow-cli-linux-amd64.tar.gz"
      sha256 "${linux_amd64}"
    end
  end

  def install
    bin.install "dataflow-cli"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/dataflow-cli --version")
  end
end
EOF
