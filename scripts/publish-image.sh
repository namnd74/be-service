#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 IMAGE TAG SOURCE_SHA VERSION RUN_URL RELEASE_JSON" >&2
  exit 2
}

[[ $# -eq 6 ]] || usage
IMAGE=$1
TAG=$2
SOURCE_SHA=$3
VERSION=$4
RUN_URL=$5
RELEASE_JSON=$6

[[ "$SOURCE_SHA" =~ ^[0-9a-f]{40}$ ]] || { echo "SOURCE_SHA must be a full commit SHA" >&2; exit 1; }
[[ "$TAG" == sha-* ]] || { echo "TAG must be a full-sha tag (sha-...)" >&2; exit 1; }
[[ "${TAG#sha-}" == "$SOURCE_SHA" ]] || { echo "TAG must contain the full source SHA" >&2; exit 1; }
[[ -n "$IMAGE" && "$IMAGE" != *[[:space:]]* ]] || { echo "IMAGE must be non-empty and contain no whitespace" >&2; exit 1; }
[[ "$VERSION" =~ ^v[0-9]+[.][0-9]+[.][0-9]+$ ]] || { echo "VERSION must match vX.Y.Z" >&2; exit 1; }
[[ "$RUN_URL" =~ ^https?://[^[:space:]]+$ ]] || { echo "RUN_URL must be an http(s) URL" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "jq is required to write release metadata safely" >&2; exit 1; }

REF="${IMAGE}:${TAG}"
echo "Pushing $REF"
docker push "$REF"

# Query the registry manifest after push; local image IDs are not registry digests.
DIGEST=$(docker buildx imagetools inspect "$REF" --format '{{.Manifest.Digest}}')
if [[ ! "$DIGEST" =~ ^sha256:[0-9a-f]{64}$ ]]; then
  echo "registry returned an invalid manifest digest: $DIGEST" >&2
  exit 1
fi

jq -n \
  --arg image "$IMAGE" \
  --arg digest "$DIGEST" \
  --arg source_sha "$SOURCE_SHA" \
  --arg version "$VERSION" \
  --arg source_run_url "$RUN_URL" \
  '{image:$image, digest:$digest, source_sha:$source_sha, version:$version, source_run_url:$source_run_url}' \
  > "$RELEASE_JSON"

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  {
    echo "image_digest=$DIGEST"
    echo "image_ref=$REF"
    echo "release_json=$RELEASE_JSON"
    echo "source_run_url=$RUN_URL"
  } >> "$GITHUB_OUTPUT"
fi
echo "Published manifest: $DIGEST"
