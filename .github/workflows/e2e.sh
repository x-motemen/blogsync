#!/bin/bash

set -e

check_variables() {
  test -n "$BLOGSYNC_BLOG"     || { echo "BLOGSYNC_BLOG is not set" >&2;     exit 1; }
  test -n "$BLOGSYNC_OWNER"    || { echo "BLOGSYNC_OWNER is not set" >&2;    exit 1; }
  test -n "$BLOGSYNC_USERNAME" || { echo "BLOGSYNC_USERNAME is not set" >&2; exit 1; }
  test -n "$BLOGSYNC_PASSWORD" || { echo "BLOGSYNC_PASSWORD is not set" >&2; exit 1; }
}

setup_config() {
  cat > blogsync.yaml <<__YAML__
$BLOGSYNC_BLOG:
  owner: $BLOGSYNC_OWNER
__YAML__
}

pull_entries() {
  blogsync pull "$BLOGSYNC_BLOG"
}

post_new_entry() {
  blogsync post --custom-path "$entry_id" "$BLOGSYNC_BLOG" <<__MARKDOWN__
---
Title: CI entry $entry_id
---

- GITHUB_SHA: \`$GITHUB_SHA\`
- GITHUB_RUN_ID: \`$GITHUB_RUN_ID\`
__MARKDOWN__
}

entry_id=$(date +%s.%N)

check_variables;

setup_config;

pull_entries;
entry_files=$(find "$BLOGSYNC_BLOG"/entry | sort)

post_new_entry;

pull_entries;
entry_files_2=$(find "$BLOGSYNC_BLOG"/entry | sort)

diff=$(diff <(echo "$entry_files") <(echo "$entry_files_2") || true)

if echo "$diff" | grep -F "> $BLOGSYNC_BLOG/entry/$entry_id.md"; then
  echo "Entry $entry_id was posted successfully"
else
  exit 1
fi